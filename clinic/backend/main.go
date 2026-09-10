package main

import (
    "log"
    "math/rand"
    "net/http"
    "os"
    "time"
)

func main() {
    s := &Server{store: NewStore()}

    mux := http.NewServeMux()

    // ---- auth ----
    mux.HandleFunc("POST /api/auth/register", s.handleRegister)
    mux.HandleFunc("POST /api/auth/login", s.handleLogin)
    mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
    mux.HandleFunc("GET /api/me", s.auth(s.handleMe))

    // ---- patient-facing ----
    mux.HandleFunc("GET /api/me/appointments", s.authRole("user", s.handleMyAppointments))
    mux.HandleFunc("POST /api/me/appointments", s.authRole("user", s.handleBookAppointment))
    mux.HandleFunc("POST /api/me/appointments/{id}/cancel", s.authRole("user", s.handleCancelMyAppointment))
    mux.HandleFunc("GET /api/slots", s.auth(s.handleSlots))
    mux.HandleFunc("GET /api/notifications", s.auth(s.handleNotifications))

    // ---- staff: appointments ----
    mux.HandleFunc("GET /api/appointments", s.authRole("staff", s.handleListAppointments))
    mux.HandleFunc("POST /api/appointments", s.authRole("staff", s.handleCreateAppointment))
    mux.HandleFunc("PUT /api/appointments/{id}/status", s.authRole("staff", s.handleSetAppointmentStatus))
    mux.HandleFunc("DELETE /api/appointments/{id}", s.authRole("staff", s.handleDeleteAppointment))
    mux.HandleFunc("POST /api/appointments/{id}/attachment", s.authRole("staff", s.handleUploadAttachment))

    // ---- files ----
    mux.HandleFunc("GET /api/files/{id}", s.handleDownloadFile)

    // ---- staff: patients ----
    mux.HandleFunc("GET /api/patients", s.authRole("staff", s.handleListPatients))

    // ---- clinic configuration ----
    mux.HandleFunc("GET /api/clinic/hours", s.handleGetHours)
    mux.HandleFunc("PUT /api/clinic/hours", s.authRole("staff", s.handlePutHours))
    mux.HandleFunc("GET /api/clinic/branding", s.handleGetBranding)
    mux.HandleFunc("PUT /api/clinic/branding", s.authRole("staff", s.handlePutBranding))
    mux.HandleFunc("POST /api/clinic/branding/logo", s.authRole("staff", s.handleUploadLogo))

    // ---- staff: message templates ----
    mux.HandleFunc("GET /api/templates", s.authRole("staff", s.handleListTemplates))
    mux.HandleFunc("POST /api/templates", s.authRole("staff", s.handleCreateTemplate))
    mux.HandleFunc("PUT /api/templates/{id}", s.authRole("staff", s.handleUpdateTemplate))
    mux.HandleFunc("DELETE /api/templates/{id}", s.authRole("staff", s.handleDeleteTemplate))

    // ---- staff: dashboard ----
    mux.HandleFunc("GET /api/stats", s.authRole("staff", s.handleStats))
    mux.HandleFunc("GET /api/stats/live", s.authRole("staff", s.handleStatsLive))
    mux.HandleFunc("GET /api/activity", s.authRole("staff", s.handleActivity))

    // ---- feedback widget ----
    mux.HandleFunc("POST /api/feedback", s.handlePostFeedback)
    mux.HandleFunc("GET /api/feedback", s.authRole("staff", s.handleListFeedback))

    // background "system" activity so the live feed keeps moving even when idle
    go func() {
        msgs := []string{
            "Automated reminder batch sent",
            "Insurance sync completed",
            "Nightly backup finished",
            "Lab results importer ran",
            "Calendar sync OK",
            "System health check passed",
        }
        for {
            time.Sleep(jitter(3500, 8000))
            s.store.AddActivity(msgs[rand.Intn(len(msgs))])
        }
    }()

    // static frontend
    staticDir := os.Getenv("STATIC_DIR")
    if staticDir == "" {
        if _, err := os.Stat("../frontend"); err == nil {
            staticDir = "../frontend"
        } else {
            staticDir = "frontend"
        }
    }
    mux.Handle("/", http.FileServer(http.Dir(staticDir)))

    log.Println("clinic-qa-lab backend listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", withCORS(mux)))
}