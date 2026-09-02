# MVP Implementation Plan - Lifeline Backend

This plan outlines the steps to build a bare-minimum viable product (MVP) for the Lifeline backend. All non-essential features (e.g., advanced logging, rate limiting, R2 file storage) are stripped out in favor of simple, functional code and core unit tests.

---

## Phase 1: Database & Model Refinements
1. **Update Models**:
   - Add `RequestLatitude` (`float64`) and `RequestLongitude` (`float64`) to `EmergencyRequest` in [`models.go`](file:///D:/Codes/LIFELINE/lifeline_backend/internal/models/models.go).
   - Define `PharmacyResponse` struct:
     ```go
     type PharmacyResponse struct {
         ID         uint      `json:"id" gorm:"primaryKey"`
         PharmacyID uint      `json:"pharmacy_id"`
         RequestID  uint      `json:"request_id"`
         Availability string  `json:"availability"` // 'YES', 'NO'
         RespondedAt time.Time `json:"responded_at"`
     }
     ```
2. **Auto-Migration**:
   - Register `PharmacyResponse` in GORM's `AutoMigrate` inside [`database.go`](file:///D:/Codes/LIFELINE/lifeline_backend/internal/database/database.go).

---

## Phase 2: Authentication & Authorization (RBAC)
1. **Repository/Service Layer**:
   - Create user creation and lookup queries.
   - Implement password hashing with `bcrypt`.
   - Implement JWT token signing and parsing helpers.
2. **Handlers & Middleware**:
   - POST `/api/v1/auth/signup`: Create a User (Patient or Pharmacy Profile).
   - POST `/api/v1/auth/login`: Validate credentials and return JWT.
   - Middleware `JWTAuth`: Verify JWT, extract claims (`UserID`, `Role`), and set them in Gin context.
   - Middleware `RequireRole`: Enforce role authorization (`PATIENT` vs `PHARMACY`).

---

## Phase 3: Emergency Requests & Location Discovery
1. **Request Creation Endpoint (Patient-only)**:
   - POST `/api/v1/requests`: Patient uploads request with latitude, longitude, note, and mock/simple prescription URL.
2. **Geospatial Discovery (Database level)**:
   - Write a raw query in repository layer to find all pharmacies within a 3km/5km radius of request coordinates.
   - Fallback/radius math using simple Haversine formula query or PostGIS `ST_DistanceSphere` depending on setup.

---

## Phase 4: WebSocket Real-Time Broadcasting
1. **WebSocket Hub**:
   - Define a `Hub` that maintains active websocket connections for connected pharmacies.
   - Handle connection upgrades (`/api/v1/ws/pharmacy`) for authenticated pharmacy accounts.
2. **Broadcaster**:
   - When a request is created, query nearby pharmacies, retrieve their connected WebSocket client channels, and broadcast the request payload to them in real time.

---

## Phase 5: Response Handling
1. **Submit Response (Pharmacy-only)**:
   - POST `/api/v1/responses` or websocket message: Pharmacy submits a response (`YES` / `NO`).
2. **Relay Update**:
   - Relate availability response to the requesting patient via socket or active dashboard polling.

---

## Phase 6: Docker Integration & Unit Testing
1. **Unit Tests**:
   - Write unit tests for the Auth Service (signup, login, JWT validation).
   - Write integration/unit tests for the geospatial query logic.
2. **Dockerization**:
   - Configure [`Dockerfile`](file:///D:/Codes/LIFELINE/lifeline_backend/Dockerfile) with a multi-stage Go build.
   - Configure [`docker-compose.yaml`](file:///D:/Codes/LIFELINE/lifeline_backend/docker-compose.yaml) to run the Go server and a PostgreSQL/PostGIS database.
