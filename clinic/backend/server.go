package main

import "net/http"

type Server struct {
    store *Store
}

// auth wraps a handler requiring any valid bearer token.
func (s *Server) auth(h http.HandlerFunc) http.HandlerFunc { return s.authRole("", h) }

// authRole wraps a handler requiring a valid token AND (if role != "") the given role.
func (s *Server) authRole(role string, h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := bearerToken(r)
        if token == "" {
            writeErr(w, http.StatusUnauthorized, "missing bearer token")
            return
        }
        userID, ok := s.store.UserIDByToken(token)
        if !ok {
            writeErr(w, http.StatusUnauthorized, "invalid or expired token")
            return
        }
        user := s.store.User(userID)
        if user == nil {
            writeErr(w, http.StatusUnauthorized, "invalid token")
            return
        }
        if role != "" && user.Role != role {
            writeErr(w, http.StatusForbidden, "forbidden: requires "+role+" role")
            return
        }
        h(w, r)
    }
}

func (s *Server) currentUser(r *http.Request) *User {
    token := bearerToken(r)
    if token == "" {
        return nil
    }
    id, ok := s.store.UserIDByToken(token)
    if !ok {
        return nil
    }
    return s.store.User(id)
}

func withCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}