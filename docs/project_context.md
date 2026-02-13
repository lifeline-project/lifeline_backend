# LifeLine Project – Backend Context File (v1 / MVP)

## 1. Project Overview

**Project Name:** LifeLine – Emergency Medicine Broadcasting Platform  
**Project Type:** College Minor → Major Project  
**Primary Goal:**  
LifeLine is a real-time platform designed to help users quickly locate nearby pharmacies that have urgent medicines available during medical emergencies. Instead of relying on static inventories or delayed delivery systems, the platform broadcasts emergency medicine requests to nearby pharmacies and collects live responses.

The core idea is to minimize the time taken for a patient to find a pharmacy that has a required medicine by using real-time communication rather than centralized inventory synchronization.

---

## 2. Problem Statement

During medical emergencies, users often struggle to locate specific medicines (e.g., insulin, injections, rare antibiotics). Existing platforms have major limitations:

- Online delivery apps are too slow for emergencies.
- Pharmacy directories (e.g., maps) do not provide real-time stock availability.
- Inventory-based systems are unreliable because small pharmacies do not maintain updated digital inventories.

As a result, users waste critical time visiting multiple pharmacies physically.

---

## 3. Proposed Solution (Core Concept)

LifeLine uses a **request–broadcast–response model**:

1. A user submits an emergency medicine request by uploading a prescription image and optional note.
2. The backend identifies nearby pharmacies using stored pharmacy locations.
3. The request is broadcast in real time to pharmacies within a defined radius.
4. Pharmacies respond with simple availability (Yes / No).
5. The user is shown pharmacies that confirmed availability.

This avoids maintaining centralized medicine inventories and focuses purely on real-time discovery.

---

## 4. MVP Scope (Version 1 – Backend-Focused)

### Included in V1 (MVP)

- User and pharmacy registration
- Authentication using email & password
- Role-based access control (Patient / Pharmacy)
- Pharmacy profile with stored location (latitude, longitude)
- Emergency request creation (prescription image + optional note)
- Real-time broadcasting of requests to pharmacies (WebSockets)
- Pharmacy response handling (Yes / No)
- Storing requests and responses in database
- Basic geospatial filtering (nearby pharmacies)

### Explicitly Excluded from V1

- Chat between user and pharmacy
- Inventory management
- Ratings / gamification
- Notifications via email / SMS / WhatsApp
- User saved addresses or location history
- Delivery or payments
- Mobile app (web-first MVP)

These are future scope (V2 / V3).

---

## 5. Backend Architecture (High-Level)

### Components

- **API Server (Go):**
  - Handles authentication, user management, emergency request creation, response handling.
- **WebSocket Service (Go):**
  - Broadcasts emergency requests to connected pharmacies in real time.
- **Database (PostgreSQL with PostGIS):**
  - Stores users, pharmacy profiles, emergency requests, pharmacy responses.
- **Geospatial Query Logic:**
  - Used to filter pharmacies based on proximity to the user’s current location (pharmacy locations are persisted; user location is provided at request time and not stored long-term except for request history).

Frontend (Flutter Web) is a client that consumes REST APIs and WebSocket events. Backend is a standalone service.

---

## 6. Data Model (Backend Entities)

### Core Entities

- **User**
  - id
  - name
  - email
  - password_hash
  - role (PATIENT / PHARMACY)
  - created_at

- **PharmacyProfile**
  - id
  - user_id (FK to User)
  - pharmacy_name
  - latitude
  - longitude
  - address

- **EmergencyRequest**
  - id
  - user_id (patient)
  - prescription_image_url
  - note
  - status (OPEN / CLOSED)
  - created_at
  - request_latitude
  - request_longitude

- **PharmacyResponse**
  - id
  - pharmacy_id
  - request_id
  - availability (YES / NO)
  - responded_at

---

## 7. Backend Functional Flow

1. **Registration**
   - User registers as Patient or Pharmacy.
   - If Pharmacy, pharmacy profile with location is created.

2. **Authentication**
   - User logs in.
   - JWT tokens issued for API access.
   - Role-based authorization enforced.

3. **Emergency Request**
   - Patient submits request with prescription image and note.
   - Backend stores request, including the user's current location for this request.
   - Backend performs geospatial query on pharmacy locations.

4. **Broadcast**
   - Backend broadcasts request via WebSockets to pharmacies within a 3 km radius.
   - If no response, radius expands to 5 km.
   - Pharmacies receive request in real time (only those connected via WebSocket for MVP).

5. **Response**
   - Pharmacy submits Yes / No.
   - Backend stores response.
   - Backend sends updated availability to the patient via WebSocket.

6. **Request Closure**
   - Patient can close their request.
   - Requests auto-close after a set timeout (e.g., 20 minutes).

---

## 8. Key Backend Design Principles

- Stateless REST APIs for core operations
- WebSockets for low-latency real-time updates
- Separation of concerns:
  - Auth module
  - Request module
  - Real-time communication module
- Clean API contracts between frontend and backend
- Avoid premature optimization and feature creep in V1
- Keep database schema minimal and aligned with MVP scope

---

## 9. Future Scope (Post-MVP)

(Not to be implemented now, but useful for design extensibility)

- Chat between user and pharmacy
- Notification fallback (email, WhatsApp, push notifications, Telegram)
- User saved addresses
- Ratings / trust score for pharmacies
- Inventory hints (optional)
- Mobile app client

---

## 10. Development Priorities (Backend)

1. Define API contract (endpoints + request/response schema)
2. Implement authentication & RBAC
3. Implement emergency request CRUD
4. Implement geospatial filtering for pharmacies
5. Implement WebSocket broadcasting
6. Integrate response handling
7. Logging & basic error handling
8. Integration testing with frontend

---

## 11. Non-Goals

- This is not a full-scale production healthcare system.
- No compliance or regulatory workflows are in scope.
- No delivery or payment workflows in MVP.

---

## 12. Success Criteria for MVP

- User can submit emergency request
- Nearby pharmacies receive the request in real time
- Pharmacy can respond Yes / No
- User sees live responses
- System works end-to-end with acceptable latency

---

## 13. Constraints

- College project scope
- Team is learning new tech
- Web-first MVP
- Real-time communication must be simple and stable
- System must be demo-ready, not production-grade

---

## 14. Technical Notes

- Prescription images are stored in Cloudflare R2 (S3-compatible).
- All geospatial queries use PostGIS.
- Only pharmacies connected via WebSocket receive broadcasts in MVP.
- User location is always fresh per request and not stored long-term, except for request history.

