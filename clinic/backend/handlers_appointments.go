package main

import (
    "fmt"
    "hash/fnv"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"
)

// pseudoUnavailable is deterministic per (date, slot) so slot lists are stable.
func pseudoUnavailable(date, slot string) bool {
    h := fnv.New32a()
    _, _ = h.Write([]byte(date + "|" + slot))
    return h.Sum32()%5 == 0
}

// ---------- patient-facing ----------

func (s *Server) handleMyAppointments(w http.ResponseWriter, r *http.Request) {
    u := s.currentUser(r)
    time.Sleep(jitter(300, 900))
    items := []*Appointment{}
    for _, a := range s.store.Appointments() {
        if a.PatientID == u.ID {
            items = append(items, a)
        }
    }
    writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleBookAppointment(w http.ResponseWriter, r *http.Request) {
    u := s.currentUser(r)
    var req struct {
        Date   string `json:"date"`
        Time   string `json:"time"`
        Reason string `json:"reason"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    if !validDate(req.Date) {
        writeErr(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
        return
    }
    if !validTime(req.Time) {
        writeErr(w, http.StatusBadRequest, "time must be HH:MM")
        return
    }
    if strings.TrimSpace(req.Reason) == "" {
        writeErr(w, http.StatusBadRequest, "reason is required")
        return
    }
    // NOTE: no past-date check, no slot-availability check, no duplicate check.
    a := s.store.CreateAppointment(u.ID, u.Name, req.Date, req.Time, strings.TrimSpace(req.Reason), "pending")
    s.store.AddActivity(u.Name + " booked an appointment")
    writeJSON(w, http.StatusCreated, a)
}

func (s *Server) handleCancelMyAppointment(w http.ResponseWriter, r *http.Request) {
    u := s.currentUser(r)
    id := pathID(r)
    a := s.store.Appointment(id)
    if a == nil || a.PatientID != u.ID {
        writeErr(w, http.StatusNotFound, "appointment not found")
        return
    }
    updated := s.store.SetAppointmentStatus(id, "cancelled")
    writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleSlots(w http.ResponseWriter, r *http.Request) {
    date := r.URL.Query().Get("date")
    if !validDate(date) {
        writeErr(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
        return
    }
    time.Sleep(jitter(800, 1800)) // heavy AJAX simulation

    booked := map[string]bool{}
    for _, a := range s.store.Appointments() {
        if a.Date == date && a.Status != "cancelled" {
            booked[a.Time] = true
        }
    }

    slots := []map[string]any{}
    for mins := 9 * 60; mins <= 16*60+30; mins += 30 {
        label := fmt.Sprintf("%02d:%02d", mins/60, mins%60)
        available := !booked[label] && !pseudoUnavailable(date, label)
        slots = append(slots, map[string]any{"time": label, "available": available})
    }
    writeJSON(w, http.StatusOK, map[string]any{"date": date, "slots": slots})
}

// ---------- staff ----------

func (s *Server) handleListAppointments(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    status := q.Get("status")
    date := q.Get("date")

    time.Sleep(jitter(400, 1200)) // slow listing on purpose

    all := s.store.Appointments()
    items := []*Appointment{}
    for _, a := range all {
        if status != "" && a.Status != status {
            continue
        }
        if date != "" && a.Date != date {
            continue
        }
        items = append(items, a)
    }

    pageStr, sizeStr := q.Get("page"), q.Get("pageSize")
    if pageStr == "" && sizeStr == "" {
        writeJSON(w, http.StatusOK, map[string]any{
            "items": items, "total": len(items), "page": 1, "pageSize": len(items),
        })
        return
    }

    page := atoiDefault(pageStr, 1)
    if page < 1 {
        page = 1
    }
    pageSize := atoiDefault(sizeStr, 100)
    if pageSize < 1 || pageSize > 100 {
        pageSize = 100
    }
    start := (page - 1) * pageSize
    if start > len(items) {
        start = len(items)
    }
    end := start + pageSize
    if end > len(items) {
        end = len(items)
    }
    pageItems := items[start:end]

    writeJSON(w, http.StatusOK, map[string]any{
        "items": pageItems,
        "total": len(pageItems), // should this be the total of ALL filtered items?
        "page":  page,
        "pageSize": pageSize,
    })
}

func (s *Server) handleCreateAppointment(w http.ResponseWriter, r *http.Request) {
    var req struct {
        PatientName string `json:"patientName"`
        Date        string `json:"date"`
        Time        string `json:"time"`
        Reason      string `json:"reason"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    req.PatientName = strings.TrimSpace(req.PatientName)
    req.Reason = strings.TrimSpace(req.Reason)

    if req.PatientName == "" {
        writeErr(w, http.StatusBadRequest, "patientName is required")
        return
    }
    if !validDate(req.Date) {
        writeErr(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
        return
    }
    if !validTime(req.Time) {
        writeErr(w, http.StatusBadRequest, "time must be HH:MM")
        return
    }
    if req.Reason == "" {
        writeErr(w, http.StatusBadRequest, "reason is required")
        return
    }
    // NOTE: no check that the date is in the future.

    a := s.store.CreateAppointment(0, req.PatientName, req.Date, req.Time, req.Reason, "pending")
    s.store.AddActivity("Appointment created for " + a.PatientName)
    writeJSON(w, http.StatusCreated, a)
}

func (s *Server) handleSetAppointmentStatus(w http.ResponseWriter, r *http.Request) {
    id := pathID(r)
    if s.store.Appointment(id) == nil {
        writeErr(w, http.StatusNotFound, "appointment not found")
        return
    }
    var req struct {
        Status string `json:"status"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    if strings.TrimSpace(req.Status) == "" {
        writeErr(w, http.StatusBadRequest, "status is required")
        return
    }
    // NOTE: no validation that status is a known value or a legal transition.
    a := s.store.SetAppointmentStatus(id, req.Status)
    s.store.AddActivity(fmt.Sprintf("Appointment #%d status changed to %q", id, req.Status))
    writeJSON(w, http.StatusOK, a)
}

func (s *Server) handleDeleteAppointment(w http.ResponseWriter, r *http.Request) {
    id := pathID(r)
    if s.store.DeleteAppointment(id) {
        s.store.AddActivity(fmt.Sprintf("Appointment #%d deleted", id))
        w.WriteHeader(http.StatusNoContent)
        return
    }
    writeErr(w, http.StatusNotFound, "appointment not found")
}

func (s *Server) handleUploadAttachment(w http.ResponseWriter, r *http.Request) {
    u := s.currentUser(r)
    id := pathID(r)
    if s.store.Appointment(id) == nil {
        writeErr(w, http.StatusNotFound, "appointment not found")
        return
    }
    f, hdr, err := r.FormFile("file")
    if err != nil {
        writeErr(w, http.StatusBadRequest, "multipart form field 'file' is required")
        return
    }
    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "failed to read file")
        return
    }
    if len(data) == 0 {
        writeErr(w, http.StatusBadRequest, "uploaded file is empty")
        return
    }
    // NOTE: no file-type or file-size validation.

    rec := s.store.SaveFile(hdr.Filename, hdr.Header.Get("Content-Type"), data, u.ID)
    s.store.SetAppointmentAttachment(id, rec.ID)
    s.store.AddActivity(fmt.Sprintf("File %q attached to appointment #%d", rec.Name, id))
    writeJSON(w, http.StatusCreated, map[string]any{
        "id": rec.ID, "name": rec.Name,
        "url": "/api/files/" + strconv.Itoa(rec.ID), "size": rec.Size,
    })
}

func (s *Server) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
    // NOTE: this endpoint intentionally has NO auth — should it?
    rec := s.store.File(pathID(r))
    if rec == nil {
        writeErr(w, http.StatusNotFound, "file not found")
        return
    }
    ct := rec.ContentType
    if ct == "" {
        ct = "application/octet-stream"
    }
    w.Header().Set("Content-Type", ct)
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", rec.Name))
    _, _ = w.Write(rec.Data)
}