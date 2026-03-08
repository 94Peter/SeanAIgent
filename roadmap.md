# Sean AIgent Project Roadmap

## 🚀 v2.0.0 - Performance & Event-Driven Foundation (Released)
*Released: 2026-02-19 | Refined: 2026-03-07*

### ✅ Infrastructure & Architecture (Completed)
- **Event-Driven Core**: Implemented internal `EventBus` and MongoDB `EventStore` for asynchronous decoupled workflows.
- **High-Performance Booking**: 
    - Achieved **4,498 RPS** with **2.16ms** avg latency.
    - Implemented `SingleFlight` to prevent cache stampedes.
    - Optimized `groupToWeeks` algorithm to $O(N)$.
- **Asynchronous Cache Invalidation**: Background `CacheWorker` pool for non-blocking cleanup.
- **Observability**: Optimized OpenTelemetry sampling (5%) for reduced GC overhead.

---

## 🛠️ v2.1.0 - Dual-Track Optimization (In Progress)

### Track A: Client-Facing Booking UX (預約端優化)
*Goal: Provide a seamless, "app-like" booking experience for students.*

- [x] **Infinite Scrolling Calendar**: Week-by-week loading with stable anchoring.
- [x] **Student Stats Dashboard**: 90-day overview of attendance and bookings.
- [ ] **Frontend Resource Separation**: Extract inline JS from `.templ` to `/assets/js/`.
- [ ] **Native Toast Integration**: Replace `alert()` with project-native Toast system.
- [ ] **Skeleton Screens**: Add loading placeholders for smoother infinite scroll transitions.
- [ ] **Input Robustness**: Enhanced server-side validation for participant names.

### Track B: Coach & Admin Analytics (教練管理後台)
*Goal: Data-driven management with automated attendance tracking.*

- [x] **Pre-aggregated Analytics (Snapshot)**: Implementation of `UserMonthlyStat` for instant report loading.
- [x] **Event-Driven Stats Linkage**: Automatic updates via `AppointmentStatusChanged` events.
- [x] **Batch Attendance Updates**: Bulk marking of attendance/absence in admin UI.
- [x] **Automated Cron Jobs**: Nightly sync for "Auto Mark Absent" and stats calibration.
- [x] **Server-side Report Engine**: Pagination and search for student attendance reports.
- [ ] **Member Billing & Payment Tracking**: Tracking student payment status (paid/unpaid/expired) and history.
- [ ] **Advanced Data Visualization**: Charts for revenue trends and class occupancy.
- [ ] **CSV Export Enhancements**: Flexible date ranges and filters for financial reconciliation.

### Track C: Tech Debt & Infrastructure (技術債與底層維護)
*Goal: Maintain a clean, maintainable codebase and efficient database.*

- [ ] **Database V1 Cleanup**: Remove legacy V1 fields from MongoDB collections and Domain Entities to save storage and reduce complexity.

---

## 🔮 v2.2.0+ - Future Considerations
- [ ] **Real-time Notification Center**: In-app/LINE notifications for booking approvals and class changes.
- [ ] **Multi-language Expansion**: Full i18n coverage for both Client and Admin UIs.
- [ ] **Waitlist System**: Automated queue management for fully booked classes.
- [ ] **Payment Integration**: Support for package purchases and credit management.

---
*Roadmap last updated: 2026-03-07 by Gemini CLI*
