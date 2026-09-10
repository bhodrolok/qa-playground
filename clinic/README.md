# Clinic QA Lab

A deliberately _imperfect_ clinic-management web app + mock REST API, built as a practice target for Selenium WebDriver, TestNG, REST Assured, and Cucumber.

Idea: Front-end (plus the API) is testing framework-agnostic

# Stack

Frontend: plain HTML/CSS/JS (no build step)
Backend: Go (standard library only — no external dependencies)
Storage: in-memory (resets on backend restart — good for repeatable runs)

# Assembly and running

Requirements: Go 1.22 or newer from the [official site](https://go.dev/doc/install) or through whatever package manager's used by your operating system.

1. `cd` into the [`clinic/backend`](./backend/) directory in a terminal 
2. Run: `go run .` 

Then open: http://localhost:8080 on a web browser.

The server auto-detects the [`frontend/`](./frontend/) folder relative to itself. You can override it with: `STATIC_DIR=/custom/path/to/frontend go run .`

# Seeded accounts

| Role | Email | Password |
|---|---|---|
| staff | staff@clinic.com | Password123 |
| user	| patient@clinic.com | Password123 |

# Intentional behavior

- The global search bar in the staff top nav is decorative (non-functional by design).
- "Live metrics" values on the staff dashboard are randomized on every refresh.
- Several endpoints have artificial latency (so you must use proper waits):
    - /api/stats ~1.5–3s
    - /api/stats/live ~0.8–2s
    - /api/slots ~0.8–1.8s
    - /api/appointments (GET) ~0.4–1.2s
    - /api/notifications ~0.6–1.5s
- Data resets when the backend restarts.
- The application intentionally contains a number of seeded defects for you todiscover. 

# Areas to practice!

| Area | Location / Example |
|---|---|
| Login / register / dashboards | `index.html`, `register.html`, `staff/*`, `user/dashboard.html` |
| Dynamic elements & heavy AJAX | Staff dashboard (skeleton stats, live metrics, activity feed), booking flow |
| Waits (visibility/invisibility) | Skeleton loaders, spinners, toasts (auto-dismiss ~4s) |
| Native browser alert | Staff → Appointments → Delete |
| Custom inline confirm (2 clicks) | Staff → Message Templates → Delete |
| File upload | Staff → Appointments (per-row "Attach"), Staff → Branding (logo) |
| File download | Appointment attachment "Download" link |
| Nested iframes | User dashboard → "Clinic notices" iframe → inner feedback iframe |
| Shadow DOM | The feedback widget inside the innermost iframe (open shadow root) |
| Chained waits | User booking: pick date → spinner → slots render → select → Book enabled |
| Stale elements | Activity feed (items added/removed continuously) |

# API summary

Auth: Authorization: Bearer <token> (token from login).

Method	Path	Auth	Purpose
POST	/api/auth/register	public	Register (name,email,password,role)
POST	/api/auth/login	public	Login → token + user
POST	/api/auth/logout	bearer	Invalidate token
GET	/api/me	bearer	Current profile
GET	/api/me/appointments	user	My appointments
POST	/api/me/appointments	user	Book appointment
POST	/api/me/appointments/{id}/cancel	user	Cancel my appointment
GET	/api/slots?date=YYYY-MM-DD	bearer	Available slots (slow)
GET	/api/notifications	bearer	Notifications (slow)
GET	/api/appointments?page=&pageSize=&status=&date=	staff	List (supports filters/pagination)
POST	/api/appointments	staff	Create appointment
PUT	/api/appointments/{id}/status	staff	Change status
DELETE	/api/appointments/{id}	staff	Delete appointment
POST	/api/appointments/{id}/attachment	staff	Multipart upload ("file" field)
GET	/api/files/{id}	?	Download file
GET	/api/patients	staff	Patient list
GET	/api/clinic/hours	public	Clinic hours
PUT	/api/clinic/hours	staff	Update hours
GET	/api/clinic/branding	public	Branding settings
PUT	/api/clinic/branding	staff	Update branding
POST	/api/clinic/branding/logo	staff	Multipart logo upload
GET	/api/templates	staff	List templates
POST	/api/templates	staff	Create template
PUT	/api/templates/{id}	staff	Update template
DELETE	/api/templates/{id}	staff	Delete template
GET	/api/stats	staff	Dashboard stats (slow)
GET	/api/stats/live	staff	Live metrics (slow, randomized)
GET	/api/activity?after={id}	staff	Activity feed (poll)
POST	/api/feedback	public	Submit feedback (widget)
GET	/api/feedback	staff	List feedback