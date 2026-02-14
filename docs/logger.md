# Logger Guide for LifeLine Backend

This guide explains how to use the logging system in the LifeLine backend, including setup, usage, log levels, and best practices. It merges requirements and practical usage for developers.

---

## 1. Logging Stack Overview

- **Backend:** Go (Gin/Fiber/net/http)
- **Logger:** Uber Zap (structured JSON logs)
- **Log Aggregation:** Grafana Loki (Grafana Cloud – Free Tier)
- **Log Shipping Agent:** Grafana Agent or Promtail
- **Visualization:** Grafana Cloud UI

---

## 2. Logging Requirements & Best Practices

- All logs must be in JSON format in production (pretty in dev).
- Use Uber Zap in production mode for structured logs.
- Include these base fields in every log:
  - `service` (e.g., "api", "auth")
  - `env` (dev/prod)
  - `level` (info/warn/error)
  - `timestamp`
- Log only important events:
  - Server start/stop
  - DB connection success/failure
  - External API failures
  - Auth failures
  - Panics / unexpected errors
- Do **not** log sensitive data (user_id, email, tokens) as Loki labels. These can be in the log body.
- Loki labels should be low-cardinality: service, env, instance.
- Application logs go to stdout or log files for shipping to Loki.
- Use human-readable logs in development, JSON logs in production.

---

## 3. Log Levels Explained

- **DEBUG:** Detailed information for debugging. Use only in development.
- **INFO:** Routine lifecycle and business events (e.g., server started, user logged in).
- **WARN:** Abnormal but recoverable events (e.g., retrying connection).
- **ERROR:** Failures that need attention (e.g., DB connection failed).
- **FATAL:** Critical errors causing shutdown (e.g., cannot start server).

**When to use which level:**
- Use `DEBUG` for development-only details.
- Use `INFO` for normal operations and successful events.
- Use `WARN` for unexpected but non-fatal issues.
- Use `ERROR` for actual failures.
- Use `FATAL` for unrecoverable errors (app will exit).

---

## 4. How to Use the Logger

### Import the logger
```go
import "lifeline_backend/pkg/logger"
import "go.uber.org/zap"
```

### Log messages
> **Note:** For details on how to add structured fields to your logs (like zap.String, zap.Error, etc.), see [Understanding Zap Fields](#5-understanding-zap-fields-zapstring-zaperror-zapany-etc).

```go
logger.Logger.Info("User registered", zap.String("email", email))
logger.Logger.Warn("Rate limit exceeded", zap.String("ip", ip))
logger.Logger.Error("DB error", zap.Error(err))
logger.Logger.Debug("Request details", zap.Any("request", req))
logger.Logger.Fatal("Startup failed", zap.Error(err))
```

### Add structured fields
Use Zap fields for context:
```go
logger.Logger.Info("User login", zap.String("email", email), zap.String("ip", ip))
```

### Example in a handler
```go
func homeHandler(w http.ResponseWriter, r *http.Request) {
    logger.Logger.Info("Home handler called", zap.String("method", r.Method), zap.String("path", r.URL.Path))
    fmt.Fprintln(w, "Welcome to LifeLine Backend!")
}
```

---

## 5. Understanding Zap Fields (zap.String, zap.Error, zap.Any, etc.)

Zap provides helper functions to add structured fields to your logs. These fields make your logs more useful and queryable.

### Common Zap Field Functions

- `zap.String(key, value)`
  - Adds a string field to the log (e.g., email, IP address).
  - Example: `zap.String("email", email)`
- `zap.Int(key, value)`
  - Adds an integer field (e.g., user ID, status code).
  - Example: `zap.Int("user_id", userID)`
- `zap.Bool(key, value)`
  - Adds a boolean field (e.g., success, isActive).
  - Example: `zap.Bool("success", true)`
- `zap.Error(err)`
  - Adds an error field. Zap will log the error message under the key "error".
  - Example: `zap.Error(err)`
- `zap.Any(key, value)`
  - Adds any Go value (struct, map, slice, etc.). Zap will try to encode it as JSON.
  - Example: `zap.Any("request", reqStruct)`

### Why Use Structured Fields?
- They make logs machine-readable and easy to filter/search in tools like Grafana Loki.
- You can add as many fields as needed to provide context for each log entry.

### Example Usage
```go
logger.Logger.Info("User login", zap.String("email", email), zap.String("ip", ip))
logger.Logger.Error("DB error", zap.Error(err))
logger.Logger.Debug("Request details", zap.Any("request", req))
```

**Summary:**
- Use `zap.String` for text values, `zap.Int` for numbers, `zap.Error` for errors, and `zap.Any` for complex data.
- Always add relevant context to your logs for better debugging and observability.

---

## 6. Local vs Production Setup

- **Development:**
  - Human-readable logs (Zap dev logger)
  - Logs printed to console
- **Production:**
  - JSON logs
  - Shipped to Grafana Cloud Loki
  - Viewable in Grafana UI

---

## 7. Observability & Loki

- Logs are queryable in Grafana using LogQL.
- Filter logs by service, log level, error messages.
- Optionally include `request_id` in log body for tracing.
- Avoid high-cardinality labels to stay within Grafana Cloud Free Tier limits.

---

## 8. Outcome

- The Go backend produces structured logs.
- Logs are centralized in Grafana Cloud Loki.
- Engineers can debug issues via Grafana dashboards.
- The logging system is production-aligned and resume-worthy.

---

## 9. How to Ship Logs to Loki

- **Grafana Agent:**
  - Install and configure the Grafana Agent.
  - Use the `loki` exporter to ship logs.
- **Promtail:**
  - Install and configure Promtail.
  - Use the `loki` exporter to ship logs.

---

## 10. Common Issues & Debugging

- **No logs appearing in Loki:**
  - Check if the agent is running and configured correctly.
  - Verify the log path and format.
- **High cardinality labels:**
  - Avoid using fields with many unique values.
  - Use low-cardinality fields like `service`, `env`, `instance`.

---

## 11. Summary

The Go backend produces structured logs, which are shipped to Grafana Cloud Loki for centralized logging and observability. Engineers can debug issues via Grafana dashboards, and the logging system is production-aligned and resume-worthy.

---

## 12. References

- [Uber Zap Documentation](https://github.com/uber-go/zap)
- [Grafana Loki Documentation](https://grafana.com/docs/loki/)

---
