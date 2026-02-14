# Lifeline Backend

Lifeline is a real-time emergency medicine availability platform. The backend helps users quickly locate nearby pharmacies that have required medicines in stock by broadcasting emergency requests and collecting responses in real time. This service will power the mobile and web clients and handle realtime communication, persistence, and API access.

Project status
We are in the early setup phase. The core infrastructure is working and the server starts with configuration, logging, database connectivity, and readiness checks. Feature development begins from this foundation.

What this project presents
- A backend API for emergency medicine availability.
- Realtime request broadcasting to nearby pharmacies.
- A foundation for inventory responses, user requests, and pharmacy participation.
- Operational readiness with health checks, logging, and clean shutdown.

Features completed so far

Core infrastructure
- Server startup with routing and middleware support.
- Health and readiness endpoints for deployments.
- Graceful shutdown to close HTTP and database cleanly.

Configuration and logging
- Environment-based configuration with defaults and validation.
- Centralized logging with configurable level and format.
- Request logging middleware for observability.

Database setup
- PostgreSQL connection with pooling and startup ping.
- Optional auto-migration for MVP development.

Current endpoints
- GET / returns a welcome message.
- GET /healthz returns liveness status.
- GET /readyz returns readiness status with a database ping.

Features planned next

Core API features
- Build endpoints for users, pharmacies, and emergency requests.
- Define request validation and response DTOs for each domain.

Authentication and authorization
- Add token-based authentication for users and pharmacies.
- Protect routes and enforce role-based access.

Domain layering
- Implement repository and service layers for clean business logic.
- Add unit tests for services and integration tests for repositories.

Realtime workflows
- Add websocket or push-based messaging for broadcasts and responses.
- Track delivery status and response time metrics.

Operations and migrations
- Introduce explicit migrations when schema changes grow.
- Extend logging conventions and health checks for production readiness.