package main

import (
    "net/http"
    "regexp"
    "strings"
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name     string `json:"name"`
        Email    string `json:"email"`
        Password string `json:"password"`
        Role     string `json:"role"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    req.Name = strings.TrimSpace(req.Name)
    req.Email = strings.TrimSpace(req.Email)

    if req.Name == "" {
        writeErr(w, http.StatusBadRequest, "name is required")
        return
    }
    if !emailRe.MatchString(req.Email) {
        writeErr(w, http.StatusBadRequest, "email is not valid")
        return
    }
    if len(req.Password) < 8 {
        writeErr(w, http.StatusBadRequest, "password must be at least 8 characters")
        return
    }

    role := "user"
    if req.Role == "staff" {
        role = "staff" // anyone may self-register as staff — is that OK?
    }

    if s.store.EmailExistsExact(req.Email) {
        writeErr(w, http.StatusConflict, "email already registered")
        return
    }

    u := s.store.CreateUser(req.Name, req.Email, req.Password, role)
    s.store.AddActivity("New account registered: " + u.Email)
    writeJSON(w, http.StatusCreated, u)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := readJSON(r, &req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    u := s.store.FindUserByEmail(strings.TrimSpace(req.Email))
    if u == nil || u.Password != hashPassword(req.Password) {
        writeErr(w, http.StatusUnauthorized, "invalid email or password")
        return
    }
    token := s.store.CreateToken(u.ID)
    s.store.AddActivity(u.Email + " signed in")
    writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
    if token := bearerToken(r); token != "" {
        s.store.DeleteToken(token)
    }
    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
    u := s.currentUser(r)
    if u == nil {
        writeErr(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    writeJSON(w, http.StatusOK, u)
}