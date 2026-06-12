# Live Commerce BI API Documentation

**Base URL**: `http://localhost:8080/api/v1`
**Authentication**: Bearer Token (JWT)
**Version**: 2.0.0

---

## Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/login` | User login |
| POST | `/auth/register` | Register new user |
| GET | `/auth/profile` | Get current user profile *(Auth)* |

### Login Request
```json
{
  "username": "string",
  "password": "string"
}
```

### Login Response
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2024-01-15T12:00:00Z",
  "user": { "id": 1, "username": "admin", "role": "admin" }
}
```

---

## Dashboard

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/dashboard/overview` | Dashboard overview statistics |
| GET | `/dashboard/realtime` | Real-time metrics |
| GET | `/dashboard/trend` | Trend data for charts |

### Custom Dashboards

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/dashboards` | Create dashboard |
| GET | `/dashboards` | List dashboards |
| GET | `/dashboards/:id` | Get dashboard |
| PUT | `/dashboards/:id` | Update dashboard |
| DELETE | `/dashboards/:id` | Delete dashboard |
| POST | `/dashboards/:id/widgets` | Add widget |
| GET | `/dashboards/:id/widgets` | List widgets |
| PUT | `/dashboards/:id/widgets/:wid` | Update widget |
| DELETE | `/dashboards/:id/widgets/:wid` | Delete widget |

---

## Data Sources *(Admin, Analyst)*

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/datasources` | Create data source |
| GET | `/datasources` | List data sources |
| GET | `/datasources/:id` | Get data source |
| PUT | `/datasources/:id` | Update data source |
| DELETE | `/datasources/:id` | Delete data source |
| POST | `/datasources/:id/sync` | Trigger sync |
| GET | `/datasources/:id/status` | Get sync status |

**Supported Platforms**: `douyin`, `kuaishou`, `taobao_live`, `jd_live`, `pdd_live`

---

## Live Rooms

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/live-rooms` | Create live room |
| GET | `/live-rooms` | List live rooms |
| GET | `/live-rooms/:id` | Get live room |
| GET | `/live-rooms/:id/metrics` | Get room metrics |
| GET | `/live-rooms/:id/products` | Get room products |
| GET | `/live-rooms/:id/funnel` | Get conversion funnel |

---

## Streamers

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/streamers` | Create streamer |
| GET | `/streamers` | List streamers |
| GET | `/streamers/:id` | Get streamer |
| PUT | `/streamers/:id` | Update streamer |
| GET | `/streamers/:id/performance` | Performance stats |
| GET | `/streamers/rankings` | Streamer rankings |

---

## Products

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/products` | Create product |
| GET | `/products` | List products |
| GET | `/products/:id` | Get product |
| PUT | `/products/:id` | Update product |
| GET | `/products/rankings` | Product rankings |

---

## Orders

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/orders` | List orders |
| GET | `/orders/:id` | Get order |
| GET | `/orders/stats` | Order statistics |
| GET | `/orders/revenue` | Revenue data |

---

## Analytics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/analytics/gmv` | GMV analytics |
| GET | `/analytics/conversion` | Conversion analysis |
| GET | `/analytics/platform-comparison` | Platform comparison |
| GET | `/analytics/category-analysis` | Category analysis |
| GET | `/analytics/time-analysis` | Time-based analysis |
| GET | `/analytics/funnel-analysis` | Funnel analysis |

## Viewer Analytics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/viewers/realtime` | Real-time viewer data |
| GET | `/viewers/demographics` | Viewer demographics |
| GET | `/viewers/engagement` | Viewer engagement |
| GET | `/viewers/retention` | Viewer retention |

---

## Advanced Analytics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/advanced-analytics/cohort` | Cohort retention analysis |
| GET | `/advanced-analytics/rfm` | RFM customer segmentation |
| GET | `/advanced-analytics/forecast` | Sales forecast (exp. smoothing) |
| GET | `/advanced-analytics/anomaly` | Anomaly detection (Z-score) |
| GET | `/advanced-analytics/user-path` | User behavior path |
| GET | `/advanced-analytics/engagement-heatmap` | Engagement heatmap |
| POST | `/advanced-analytics/olap` | OLAP multi-dimensional query |
| GET | `/advanced-analytics/drill-down` | Drill-down analysis |
| GET | `/advanced-analytics/period-comparison` | YoY/MoM/WoW comparison |
| GET | `/advanced-analytics/target-comparison` | Target vs actual |

### OLAP Query Body
```json
{
  "dimensions": ["platform", "date"],
  "metrics": ["gmv", "orders", "views"],
  "filters": { "platform": "douyin" },
  "limit": 100
}
```

---

## Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/reports` | Create report |
| GET | `/reports` | List reports |
| GET | `/reports/:id` | Get report |
| PUT | `/reports/:id` | Update report |
| DELETE | `/reports/:id` | Delete report |
| POST | `/reports/:id/generate` | Generate report |
| GET | `/reports/:id/download` | Download report |
| GET | `/report-templates` | List templates |
| POST | `/report-templates` | Create template |

---

## Alerts

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/alerts` | Create alert rule |
| GET | `/alerts` | List alert rules |
| GET | `/alerts/:id` | Get alert rule |
| PUT | `/alerts/:id` | Update alert rule |
| DELETE | `/alerts/:id` | Delete alert rule |
| GET | `/alerts/:id/history` | Alert history |
| POST | `/alerts/:id/test` | Test alert rule |

---

## Organization *(Admin Only)*

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/organizations` | Create organization |
| GET | `/organizations` | List organizations |
| GET | `/organizations/:id` | Get organization |
| PUT | `/organizations/:id` | Update organization |
| POST | `/organizations/:id/departments` | Create department |
| GET | `/organizations/:id/departments/tree` | Department tree |
| PUT | `/organizations/:id/departments/:did` | Update department |
| DELETE | `/organizations/:id/departments/:did` | Delete department |

---

## RBAC *(Admin Only)*

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/rbac/permissions` | List all permissions |
| POST | `/rbac/roles` | Create role |
| GET | `/rbac/roles` | List roles |
| PUT | `/rbac/roles/:id` | Update role |
| POST | `/rbac/roles/:id/users/:uid` | Assign role |
| GET | `/rbac/me/permissions` | Current user permissions |

**23 Predefined Permissions**: dashboard.view, dashboard.manage, datasource.view, datasource.manage, livestream.view, livestream.manage, streamer.view, streamer.manage, product.view, product.manage, order.view, order.manage, analytics.view, analytics.export, report.view, report.manage, alert.view, alert.manage, organization.manage, rbac.manage, audit.view, export.manage, quality.manage

---

## Audit Logs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/audit-logs` | List audit logs |
| GET | `/audit-logs/stats` | Audit statistics |

---

## Data Export

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/exports` | Create export task |
| GET | `/exports` | List exports |
| GET | `/exports/:id` | Export status |
| GET | `/exports/:id/download` | Download export |
| POST | `/exports/:id/cancel` | Cancel export |

---

## Data Quality *(Admin, Analyst)*

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/data-quality/rules` | Create quality rule |
| GET | `/data-quality/rules` | List quality rules |
| POST | `/data-quality/rules/:id/check` | Run quality check |
| GET | `/data-quality/rules/:id/results` | Check results |

---

## Event Tracking *(Public SDK)*

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/events/track` | Track single event |
| POST | `/events/batch` | Track batch events |
| GET | `/events/funnel` | Funnel events *(Auth)* |
| GET | `/events/user-paths` | User paths *(Auth)* |

---

## Aggregated Metrics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/metrics/hourly` | Hourly aggregated metrics |
| GET | `/metrics/daily` | Daily aggregated metrics |
| GET | `/metrics/streamer/:id/daily` | Streamer daily metrics |

---

## WebSocket

| Endpoint | Description |
|----------|-------------|
| `ws://host/ws?token=xxx&room_id=xxx` | WebSocket connection |
| `GET /ws/stats` | Connection statistics |
| `GET /ws/room/:room_id` | Subscribe to room |

**Real-time Channels** (via Redis Pub/Sub):
- `live:metrics:update` - Metric updates
- `live:room:alert` - Alert notifications
- `live:room:gmv` - GMV updates

---

## System

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check (DB + Redis status) |
| `GET /metrics` | System metrics |

---

## Error Response Format

```json
{
  "code": 400,
  "message": "Error description",
  "request_id": "20240115120000-a1b2c3d4"
}
```

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 429 | Rate Limited |
| 500 | Internal Server Error |
| 503 | Service Unavailable |

## Rate Limiting

- **Default**: 200 requests/minute per IP
- **Headers**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

## Pagination

All list endpoints support:
- `?page=1&page_size=20`

Response:
```json
{
  "data": [...],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

---

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│   API Server  │────▶│  PostgreSQL  │
│   (React)    │◀────│   (Go/Gin)   │     │     (16)     │
└──────────────┘     └──────┬───────┘     └──────────────┘
                           │
                     ┌─────┴─────┐
                     │           │
               ┌─────▼──┐  ┌────▼─────┐
               │ Redis  │  │  Kafka   │
               │  (7)   │  │ (Events) │
               └────────┘  └──────────┘
```

**Tech Stack**: Go 1.22, Gin, PostgreSQL 16, Redis 7, Kafka, WebSocket
