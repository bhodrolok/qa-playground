package main

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "sort"
    "strings"
    "sync"
    "time"
)

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Password  string    `json:"-"` // sha256 hex, never intended for responses
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"createdAt"`
}

type Appointment struct {
    ID           int       `json:"id"`
    PatientID    int       `json:"patientId"`
    PatientName  string    `json:"patientName"`
    Date         string    `json:"date"` // YYYY-MM-DD
    Time         string    `json:"time"` // HH:MM
    Reason       string    `json:"reason"`
    Status       string    `json:"status"` // pending | confirmed | cancelled | completed
    AttachmentID int       `json:"attachmentId,omitempty"`
    CreatedAt    time.Time `json:"createdAt"`
}

type FileRecord struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    ContentType string    `json:"contentType"`
    Size        int       `json:"size"`
    Data        []byte    `json:"-"`
    UploadedBy  int       `json:"uploadedBy"`
    CreatedAt   time.Time `json:"createdAt"`
}

type DayHours struct {
    Day    string `json:"day"`
    Open   string `json:"open"`
    Close  string `json:"close"`
    Closed bool   `json:"closed"`
}

type Branding struct {
    ClinicName   string `json:"clinicName"`
    Tagline      string `json:"tagline"`
    PrimaryColor string `json:"primaryColor"`
    LogoURL      string `json:"logoUrl"`
}

type MessageTemplate struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Channel   string    `json:"channel"`
    Subject   string    `json:"subject"`
    Body      string    `json:"body"`
    UpdatedAt time.Time `json:"updatedAt"`
}

type ActivityEvent struct {
    ID        int       `json:"id"`
    Timestamp time.Time `json:"timestamp"`
    Message   string    `json:"message"`
}

type Feedback struct {
    ID        int       `json:"id"`
    Rating    int       `json:"rating"`
    Comment   string    `json:"comment"`
    CreatedAt time.Time `json:"createdAt"`
}

type Store struct {
    mu sync.RWMutex

    users     map[int]*User
    appts     map[int]*Appointment
    files     map[int]*FileRecord
    templates map[int]*MessageTemplate

    hours    []DayHours
    branding Branding

    activity []ActivityEvent
    feedback []Feedback
    tokens   map[string]int // token -> userID

    nextUserID     int
    nextApptID     int
    nextFileID     int
    nextTemplateID int
    nextActivityID int
    nextFeedbackID int
}

func hashPassword(pw string) string {
    sum := sha256.Sum256([]byte(pw))
    return hex.EncodeToString(sum[:])
}

func newToken() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}

func NewStore() *Store {
    s := &Store{
        users:     map[int]*User{},
        appts:     map[int]*Appointment{},
        files:     map[int]*FileRecord{},
        templates: map[int]*MessageTemplate{},
        tokens:    map[string]int{},
        nextUserID: 1, nextApptID: 1, nextFileID: 1,
        nextTemplateID: 1, nextActivityID: 1, nextFeedbackID: 1,
    }
    s.seed()
    return s
}

func (s *Store) seed() {
    _ = s.CreateUser("Dr. Sarah Chen", "staff@clinic.com", "Password123", "staff")
    p1 := s.CreateUser("John Smith", "patient@clinic.com", "Password123", "user")
    _ = s.CreateUser("Maria Garcia", "maria@clinic.com", "Password123", "user")
    _ = s.CreateUser("Ahmed Khan", "ahmed@clinic.com", "Password123", "user")

    now := time.Now()
    d := func(days int) string { return now.AddDate(0, 0, days).Format("2006-01-02") }

    s.CreateAppointment(p1.ID, "John Smith", d(1), "09:30", "Annual check-up", "confirmed")
    s.CreateAppointment(p1.ID, "John Smith", d(3), "14:00", "Follow-up: lab results", "pending")
    s.CreateAppointment(0, "Walk-in Patient", d(-2), "11:00", "Flu symptoms", "completed")
    s.CreateAppointment(0, "Emily Davis", d(0), "10:30", "Vaccination", "confirmed")
    s.CreateAppointment(0, "Robert Lee", d(-1), "16:00", "Back pain consultation", "cancelled")
    s.CreateAppointment(0, "Priya Patel", d(2), "13:00", "Dental cleaning", "pending")

    s.hours = []DayHours{
        {"Monday", "09:00", "17:00", false},
        {"Tuesday", "09:00", "17:00", false},
        {"Wednesday", "09:00", "17:00", false},
        {"Thursday", "09:00", "17:00", false},
        {"Friday", "09:00", "15:00", false},
        {"Saturday", "10:00", "14:00", false},
        {"Sunday", "", "", true},
    }

    s.branding = Branding{
        ClinicName:   "Pulse Clinic",
        Tagline:      "Care that keeps up with you",
        PrimaryColor: "#2563eb",
    }

    s.CreateTemplate("Appointment Reminder (SMS)", "sms", "", "Hi {{patient_name}}, reminder: your appointment is on {{appointment_date}} at {{appointment_time}}. - {{clinic_name}}")
    s.CreateTemplate("Welcome Email", "email", "Welcome to {{clinic_name}}", "Dear {{patient_name}}, welcome! We look forward to caring for you.")
    s.CreateTemplate("Follow-up Email", "email", "How are you feeling?", "Hi {{patient_name}}, checking in after your visit on {{appointment_date}}. Reply to let us know how you're doing.")

    s.AddActivity("System seeded with demo data")
    s.AddActivity("Daily backup completed")
    s.AddActivity("Lab results importer finished")
}

// ---------- users ----------

func (s *Store) CreateUser(name, email, password, role string) *User {
    s.mu.Lock()
    defer s.mu.Unlock()
    u := &User{ID: s.nextUserID, Name: name, Email: email, Password: hashPassword(password), Role: role, CreatedAt: time.Now()}
    s.users[u.ID] = u
    s.nextUserID++
    return u
}

// FindUserByEmail is case-INSENSITIVE (used by login).
func (s *Store) FindUserByEmail(email string) *User {
    s.mu.RLock()
    defer s.mu.RUnlock()
    for _, u := range s.users {
        if strings.EqualFold(u.Email, email) {
            return u
        }
    }
    return nil
}

// EmailExistsExact is case-SENSITIVE (used by registration duplicate check).
func (s *Store) EmailExistsExact(email string) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    for _, u := range s.users {
        if u.Email == email {
            return true
        }
    }
    return false
}

func (s *Store) User(id int) *User {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.users[id]
}

func (s *Store) Users() []*User {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make([]*User, 0, len(s.users))
    for _, u := range s.users {
        out = append(out, u)
    }
    sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
    return out
}

// ---------- tokens ----------

func (s *Store) CreateToken(userID int) string {
    t := newToken()
    s.mu.Lock()
    s.tokens[t] = userID
    s.mu.Unlock()
    return t
}

func (s *Store) UserIDByToken(t string) (int, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    id, ok := s.tokens[t]
    return id, ok
}

func (s *Store) DeleteToken(t string) {
    s.mu.Lock()
    delete(s.tokens, t)
    s.mu.Unlock()
}

// ---------- appointments ----------

func (s *Store) CreateAppointment(patientID int, patientName, date, timeStr, reason, status string) *Appointment {
    s.mu.Lock()
    defer s.mu.Unlock()
    a := &Appointment{
        ID: s.nextApptID, PatientID: patientID, PatientName: patientName,
        Date: date, Time: timeStr, Reason: reason, Status: status, CreatedAt: time.Now(),
    }
    s.appts[a.ID] = a
    s.nextApptID++
    return a
}

func (s *Store) Appointments() []*Appointment {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make([]*Appointment, 0, len(s.appts))
    for _, a := range s.appts {
        out = append(out, a)
    }
    sort.Slice(out, func(i, j int) bool {
        if out[i].Date != out[j].Date {
            return out[i].Date < out[j].Date
        }
        return out[i].Time < out[j].Time
    })
    return out
}

func (s *Store) Appointment(id int) *Appointment {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.appts[id]
}

func (s *Store) SetAppointmentStatus(id int, status string) *Appointment {
    s.mu.Lock()
    defer s.mu.Unlock()
    if a, ok := s.appts[id]; ok {
        a.Status = status
        return a
    }
    return nil
}

func (s *Store) SetAppointmentAttachment(id, fileID int) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    if a, ok := s.appts[id]; ok {
        a.AttachmentID = fileID
        return true
    }
    return false
}

func (s *Store) DeleteAppointment(id int) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    if _, ok := s.appts[id]; ok {
        delete(s.appts, id)
        return true
    }
    return false
}

// ---------- files ----------

func (s *Store) SaveFile(name, contentType string, data []byte, by int) *FileRecord {
    s.mu.Lock()
    defer s.mu.Unlock()
    f := &FileRecord{
        ID: s.nextFileID, Name: name, ContentType: contentType,
        Size: len(data), Data: data, UploadedBy: by, CreatedAt: time.Now(),
    }
    s.files[f.ID] = f
    s.nextFileID++
    return f
}

func (s *Store) File(id int) *FileRecord {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.files[id]
}

// ---------- clinic config ----------

func (s *Store) Hours() []DayHours {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make([]DayHours, len(s.hours))
    copy(out, s.hours)
    return out
}

func (s *Store) SetHours(h []DayHours) {
    s.mu.Lock()
    s.hours = h
    s.mu.Unlock()
}

func (s *Store) Branding() Branding {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.branding
}

func (s *Store) SetBranding(b Branding) {
    s.mu.Lock()
    s.branding = b
    s.mu.Unlock()
}

// ---------- templates ----------

func (s *Store) Templates() []*MessageTemplate {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make([]*MessageTemplate, 0, len(s.templates))
    for _, t := range s.templates {
        out = append(out, t)
    }
    sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
    return out
}

func (s *Store) Template(id int) *MessageTemplate {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.templates[id]
}

func (s *Store) CreateTemplate(name, channel, subject, body string) *MessageTemplate {
    s.mu.Lock()
    defer s.mu.Unlock()
    t := &MessageTemplate{ID: s.nextTemplateID, Name: name, Channel: channel, Subject: subject, Body: body, UpdatedAt: time.Now()}
    s.templates[t.ID] = t
    s.nextTemplateID++
    return t
}

func (s *Store) UpdateTemplate(id int, name, channel, subject, body string) *MessageTemplate {
    s.mu.Lock()
    defer s.mu.Unlock()
    if t, ok := s.templates[id]; ok {
        t.Name, t.Channel, t.Subject, t.Body, t.UpdatedAt = name, channel, subject, body, time.Now()
        return t
    }
    return nil
}

func (s *Store) DeleteTemplate(id int) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    if _, ok := s.templates[id]; ok {
        delete(s.templates, id)
        return true
    }
    return false
}

// ---------- activity / feedback ----------

func (s *Store) AddActivity(msg string) ActivityEvent {
    s.mu.Lock()
    defer s.mu.Unlock()
    ev := ActivityEvent{ID: s.nextActivityID, Timestamp: time.Now(), Message: msg}
    s.activity = append(s.activity, ev)
    s.nextActivityID++
    return ev
}

func (s *Store) ActivityAfter(afterID int) []ActivityEvent {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := []ActivityEvent{}
    for _, ev := range s.activity {
        if ev.ID > afterID {
            out = append(out, ev)
        }
    }
    if len(out) > 50 {
        out = out[len(out)-50:]
    }
    return out
}

func (s *Store) LatestActivityID() int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if len(s.activity) == 0 {
        return 0
    }
    return s.activity[len(s.activity)-1].ID
}

func (s *Store) AddFeedback(rating int, comment string) Feedback {
    s.mu.Lock()
    defer s.mu.Unlock()
    f := Feedback{ID: s.nextFeedbackID, Rating: rating, Comment: comment, CreatedAt: time.Now()}
    s.feedback = append(s.feedback, f)
    s.nextFeedbackID++
    return f
}

func (s *Store) FeedbackList() []Feedback {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make([]Feedback, 0, len(s.feedback))
    for i := len(s.feedback) - 1; i >= 0 && len(out) < 50; i-- {
        out = append(out, s.feedback[i])
    }
    return out
}