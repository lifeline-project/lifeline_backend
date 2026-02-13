We are building a production-grade logging setup for a Go backend service. The goal is to implement **structured, centralized logging** with minimal operational overhead, suitable for a student/side project but aligned with real-world backend engineering practices.

### 🔹 Tech Stack

* Backend: Go (Gin/Fiber/net/http)
* Logger: Uber Zap (structured JSON logs)
* Log Aggregation: Grafana Loki (Grafana Cloud – Free Tier)
* Log Shipping Agent: Grafana Agent or Promtail
* Visualization: Grafana Cloud UI

---

### 🔹 Logging Requirements

1. **Structured Logging**

   * All logs must be in JSON format.
   * Use Uber Zap in production mode.
   * Include consistent base fields:

     * `service` (e.g., "api", "auth")
     * `env` (dev/prod)
     * `level` (info/warn/error)
     * `timestamp`
   * Log important events only:

     * Server start/stop
     * DB connection success/failure
     * External API failures
     * Auth failures
     * Panics / unexpected errors

2. **Log Levels**

   * DEBUG: enabled only in development
   * INFO: lifecycle and business events
   * WARN: abnormal but recoverable events
   * ERROR: failures
   * FATAL: crash scenarios

3. **No Sensitive or High-Cardinality Labels**

   * Do NOT put the following in Loki labels:

     * user_id
     * request_id
     * email
     * tokens
   * These can appear in log bodies but NOT as Loki labels.
   * Loki labels must be low-cardinality:

     * service
     * env
     * instance

4. **Log Shipping to Loki**

   * Application logs should go to stdout or log files.
   * Grafana Agent / Promtail will collect logs and push to Grafana Cloud Loki.
   * Loki endpoint should be configured using Grafana Cloud credentials (username + API key).

5. **Local vs Production Setup**

   * Development:

     * Human-readable logs (Zap dev logger)
     * Logs printed to console
   * Production:

     * JSON logs
     * Shipped to Grafana Cloud Loki
     * Viewable in Grafana UI

6. **Log Retention & Free Tier Constraints**

   * Grafana Cloud Free Tier is used (14-day retention).
   * Ensure Loki label design avoids high cardinality to stay within 10K active series limit.
   * No per-user or per-request labels.

7. **Observability**

   * Logs should be queryable in Grafana using LogQL.
   * Support filtering by:

     * service
     * log level
     * error messages
   * Optional: include request_id in log body for tracing across services.

---

### 🔹 Outcome

By the end of this setup:

* The Go backend produces structured logs.
* Logs are centralized in Grafana Cloud Loki.
* Engineers can debug issues via Grafana dashboards.
* The logging system is production-aligned and resume-worthy.

