package main

import (
    "math/rand"
    "net/http"
    "strings"
    "time"
)

// ---------- patients ----------

func (s *Server) handleListPatients(w http.ResponseWriter, r *http.Request) {
    out := []map[string]any{}
    for _, u := range s.store.Users() {
        if u.Role != "user" {
            continue
        }
        out = append(out, map[string]any{
            "id":        u.ID,
            "name":      u.Name,
            "email":     u.Email,
            "createdAt": u.CreatedAt,
            "password":  u.Password, // hmm — should this be in the response?
        })
    }
    writeJSON(w, http.StatusOK, map[string]any{"items": out, "total": len(out)})
}

// ---------- dashboard stats ----------

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
    time.Sleep(jitter(1500, 3000)) // heavy AJAX simulation

    today := time.Now().Format("2006-01-02")
    weekAgo := time.Now().AddDate(0, 0, -7)

    var total, upcoming, confirmedToday, cancelled, patients, newPatients, templates int
    for _, a := range s.store.Appointments() {
        total++
        if a.Date >= today && a.Status != "completed" { // does "upcoming" include cancelled?
            upcoming++
        }
        if a.Date == today && a.Status == "confirmed" {
            confirmedToday++
        }
        if a.Status == "cancelled" {
            cancelled++
        }
    }
    for _, u := range s.store.Users() {
        if u.Role != "user" {
            continue
        }
        patients++
        if u.CreatedAt.After(weekAgo) {
            newPatients++
        }
    }
    templates = len(s.store.Templates())

    writeJSON(w, http.StatusOK, map[string]any{
        "totalAppointments":  total,
        "upcomingAppointments": upcoming,
        "confirmedToday":     confirmedToday,
        "cancelledTotal":     cancelled,
        "totalPatients":      patients,
        "newPatientsThisWeek": newPatients,
        "activeTemplates":    templates,
    })
}

func (s *Server) handleStatsLive(w http.ResponseWriter, r *http.Request) {
    time.Sleep(jitter(800, 2000))
    writeJSON(w, http.StatusOK, map[string]any{
        "queueNow":        2 + rand.Intn(6),
        "avgWaitMinutes":  8 + rand.Intn(22),
        "onlineStaff":     1 + rand.Intn(3),
        "serverTime":      time.Now().UTC(),
    })
}

func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
    after := atoiDefault(r.URL.Query().Get("after"), 0)
    events := s.store.ActivityAfter(after)
    writeJSON(w, http.StatusOK, map[string]any{"events": events, "latest": s.store.LatestActivityID()})
}

// ---------- notifications ----------

func (s *Server) handleNotifications(w http.ResponseWriter, r *http.Request) {
    time.Sleep(jitter(600, 1500))
    items := []map[string]any{
        {"id": 1, "title": "Welcome", "body": "Thanks for choosing our clinic.", "timestamp": time.Now().Add(-time.Hour)},
        {"id": 2, "title": "Flu season", "body": "Vaccines are available on Saturdays.", "timestamp": time.Now().Add(-24 * time.Hour)},
        {"id": 3, "title": "Portal tip", "body": "You can book follow-ups from the dashboard.", "timestamp": time.Now().Add(-72 * time.Hour)},
    }
    writeJSON(w, http.StatusOK, map[string]any{"items": items, "serverTime": time.Now().UTC()})
}

// ---------- feedback (shadow-DOM widget posts here) ----------

func (s *Server) handlePostFeedback(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Rating  int    `json:"rating"`
        Comment string `json:"comment"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    // NOTE: no validation that rating is within 1..5.
    fb := s.store.AddFeedback(req.Rating, strings.TrimSpace(req.Comment))
    writeJSON(w, http.StatusCreated, map[string]any{"id": fb.ID, "message": "Thanks for your feedback!"})
}

func (s *Server) handleListFeedback(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]any{"items": s.store.FeedbackList()})
}