package main

import (
    "io"
    "net/http"
    "strconv"
    "strings"
)

var validDays = map[string]bool{
    "Monday": true, "Tuesday": true, "Wednesday": true, "Thursday": true,
    "Friday": true, "Saturday": true, "Sunday": true,
}

// ---------- hours ----------

func (s *Server) handleGetHours(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]any{"hours": s.store.Hours()})
}

func (s *Server) handlePutHours(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Hours []DayHours `json:"hours"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    if len(req.Hours) != 7 {
        writeErr(w, http.StatusBadRequest, "exactly 7 days (Monday..Sunday) are required")
        return
    }
    seen := map[string]bool{}
    for _, h := range req.Hours {
        if !validDays[h.Day] {
            writeErr(w, http.StatusBadRequest, "invalid day: "+h.Day)
            return
        }
        if seen[h.Day] {
            writeErr(w, http.StatusBadRequest, "duplicate day: "+h.Day)
            return
        }
        seen[h.Day] = true
        if !h.Closed {
            if !validTime(h.Open) || !validTime(h.Close) {
                writeErr(w, http.StatusBadRequest, "invalid time for "+h.Day+" (use HH:MM)")
                return
            }
        }
        // NOTE: no check that Open < Close.
    }
    s.store.SetHours(req.Hours)
    s.store.AddActivity("Clinic hours updated")
    writeJSON(w, http.StatusOK, map[string]any{"hours": s.store.Hours()})
}

// ---------- branding ----------

func (s *Server) handleGetBranding(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, s.store.Branding())
}

func (s *Server) handlePutBranding(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ClinicName   string `json:"clinicName"`
        Tagline      string `json:"tagline"`
        PrimaryColor string `json:"primaryColor"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    req.ClinicName = strings.TrimSpace(req.ClinicName)
    if req.ClinicName == "" {
        writeErr(w, http.StatusBadRequest, "clinicName is required")
        return
    }
    // NOTE: primaryColor is not validated as a hex color.

    b := s.store.Branding()
    b.ClinicName = req.ClinicName
    b.Tagline = strings.TrimSpace(req.Tagline)
    b.PrimaryColor = strings.TrimSpace(req.PrimaryColor)
    s.store.SetBranding(b)
    s.store.AddActivity("Clinic branding updated")
    writeJSON(w, http.StatusOK, b)
}

func (s *Server) handleUploadLogo(w http.ResponseWriter, r *http.Request) {
    u := s.currentUser(r)
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
    // NOTE: no check that the uploaded file is actually an image.

    rec := s.store.SaveFile(hdr.Filename, hdr.Header.Get("Content-Type"), data, u.ID)
    b := s.store.Branding()
    b.LogoURL = "/api/files/" + strconv.Itoa(rec.ID)
    s.store.SetBranding(b)
    s.store.AddActivity("Clinic logo updated")
    writeJSON(w, http.StatusCreated, map[string]any{"logoUrl": b.LogoURL, "name": rec.Name})
}

// ---------- templates ----------

func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
    items := s.store.Templates()
    writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

func (s *Server) handleCreateTemplate(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name    string `json:"name"`
        Channel string `json:"channel"`
        Subject string `json:"subject"`
        Body    string `json:"body"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Subject) == "" || strings.TrimSpace(req.Body) == "" {
        writeErr(w, http.StatusBadRequest, "name, subject and body are required")
        return
    }
    channel := strings.TrimSpace(req.Channel)
    if channel == "" {
        channel = "email"
    }
    t := s.store.CreateTemplate(strings.TrimSpace(req.Name), channel, strings.TrimSpace(req.Subject), req.Body)
    writeJSON(w, http.StatusCreated, t)
}

func (s *Server) handleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
    id := pathID(r)
    var req struct {
        Name    string `json:"name"`
        Channel string `json:"channel"`
        Subject string `json:"subject"`
        Body    string `json:"body"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Subject) == "" || strings.TrimSpace(req.Body) == "" {
        writeErr(w, http.StatusBadRequest, "name, subject and body are required")
        return
    }
    t := s.store.UpdateTemplate(id, strings.TrimSpace(req.Name), strings.TrimSpace(req.Channel), strings.TrimSpace(req.Subject), req.Body)
    if t == nil {
        writeErr(w, http.StatusNotFound, "template not found")
        return
    }
    writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
    id := pathID(r)
    if s.store.DeleteTemplate(id) {
        w.WriteHeader(http.StatusNoContent)
        return
    }
    writeErr(w, http.StatusNotFound, "template not found")
}