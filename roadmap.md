# Sean AIgent Project Roadmap

## 🚀 v2.0.0 - Performance & Event-Driven Foundation (Released)
*Released: 2026-02-19 | Refined: 2026-03-07*

### ✅ Infrastructure & Architecture (Completed)
- **Event-Driven Core**: Implemented internal `EventBus` and MongoDB `EventStore` for asynchronous decoupled workflows.
- **High-Performance Booking**: 
    - Achieved **4,498 RPS** with **2.16ms** avg latency.
    - Implemented `SingleFlight` to prevent cache stampedes.
- **Asynchronous Cache Invalidation**: Background `CacheWorker` pool for non-blocking cleanup.
- **Observability**: Optimized OpenTelemetry sampling (5%) for reduced GC overhead.

---

## 🛠️ v2.1.0 - Coach Backend Empowerment & UX Refinement (In Progress)
*Goal: Strengthen coach management capabilities and deliver a premium, secure student experience.*

### Track A: Client-Facing Booking UX (預約端優化)
- [x] **Infinite Scrolling Calendar**: Week-by-week loading with stable anchoring.
- [x] **Student Stats Dashboard**: 90-day overview of attendance and bookings.
- [x] **Native Toast Integration**: Replace legacy `alert()` with project-native Toast system.
- [x] **Input Robustness**: Enhanced server-side validation for participant names and contact info.
- [x] **UI Polish**: Refine spacing and interactive feedback for a more modern feel.
- [ ] **UI Animation Refinement**: Enhance transitions between calendar views and booking modals.

### Track B: Advanced Coach Backend (教練管理後台強化)
- [x] **Pre-aggregated Analytics (Snapshot)**: Implementation of `UserMonthlyStat` for instant report loading.
- [x] **Batch Attendance Updates**: Bulk marking of attendance/absence in admin UI.
- [x] **Automated Cron Jobs**: Nightly sync for "Auto Mark Absent" and stats calibration.
- [x] **Server-side Report Engine**: Pagination and search for student attendance reports.
- [x] **CSV Export**: One-click export for monthly student attendance and stats.
- [x] **Leave Reason Visibility**: Coaches can now see the student's leave reason during check-in.
- [ ] **Accounting & Payment Tracking**: View member payment records and status (Paid/Unpaid).
- [ ] **Data Visualization**: Advanced charts for revenue trends and class occupancy.

### Track C: Security Hardening (安全加固)
- [ ] **API Rate Limiting**: Implement middleware to prevent DoS on booking and export endpoints.
- [ ] **IDOR Protection**: Enforce strict horizontal permission checks (ensure users only access their own data).
- [ ] **Security Headers & CSP**: Configure CSP and HSTS to mitigate XSS and clickjacking risks.
- [x] **Go 1.25 Security Update**: Fully migrated to Go 1.25 to leverage latest security patches.

### Track D: Tech Debt & Architecture (技術債與架構維護)
- [x] **Frontend Resource Separation**: Extracted inline JS from `.templ` to `/assets/js/` for better caching.
- [x] **V1 Field Deprecation Cleanup**: Thorough removal of legacy boolean fields (`is_checked_in`, `is_on_leave`).
- [x] **Cache Busting Strategy**: Implemented versioning for all external JS assets to prevent stale code.

---

## 🏟️ v2.2.0 - Team & School Squad Management (團隊與校隊管理)
*Goal: Expand the system to handle structured groups and competitive teams with data isolation.*

### Track A: Team Management & Real-time
- [ ] **Team Creation & Member Assignment**: Capability for coaches to manage specific squads.
- [ ] **Real-time Availability Updates**: Auto-refresh slot capacity via WebSockets or long polling.
- [ ] **Role-Based Access (RBAC)**: Permission levels for head coaches vs. assistant coaches.

### Track B: School Team Attendance (校隊出缺席管理)
- [ ] **Squad Attendance Tracking**: Specialized tracking for school team practice sessions.
- [ ] **Team Scoped Security**: Ensure all records and queries are strictly isolated by `team_id`.
- [ ] **Performance Logging**: Linking attendance data with basic performance metrics or notes.

---

## 🔮 v2.3.0+ - Compliance, Auditing & Expansion
- [ ] **PII Masking**: Mask sensitive student data in non-essential admin views.
- [ ] **Automated Auditing**: Event-based dashboard for reviewing critical state changes.
- [ ] **Real-time Notification Center**: In-app/LINE notifications for booking approvals.
- [ ] **Multi-language Expansion**: Full i18n coverage for both Client and Admin UIs.

---
*Roadmap last updated: 2026-03-09 by Gemini CLI Security Engineer*
