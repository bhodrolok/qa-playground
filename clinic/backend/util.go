package main

import (
    "encoding/json"
    "math/rand"
    "net/http"
    "strconv"
    "strings"
    "time"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
    defer r.Body.Close()
    return json.NewDecoder(r.Body).Decode(v)
}

// jitter sleeps a random duration between minMs and maxMs (inclusive).
func jitter(minMs, maxMs int) time.Duration {
    return time.Duration(minMs+rand.Intn(maxMs-minMs+1)) * time.Millisecond
}

func bearerToken(r *http.Request) string {
    h := r.Header.Get("Authorization")
    if strings.HasPrefix(h, "Bearer ") {
        return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
    }
    return ""
}

func atoiDefault(s string, def int) int {
    if n, err := strconv.Atoi(s); err == nil {
        return n
    }
    return def
}

func pathID(r *http.Request) int {
    n, _ := strconv.Atoi(r.PathValue("id"))
    return n
}

func validDate(s string) bool {
    _, err := time.Parse("2006-01-02", s)
    return err == nil
}

func validTime(s string) bool {
    _, err := time.Parse("15:04", s)
    return err == nil
}