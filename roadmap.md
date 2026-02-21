# Sean AIgent Project Roadmap

## v2.0.0 - The Next Generation Booking Experience (Current)
*Released: 2026-02-19*

### ✅ Key Features Delivered
- **Infinite Scrolling Calendar**: Smooth week-by-week loading with stable scroll anchoring.
- **Student Stats Dashboard**: 90-day overview of bookings, leaves, and attendance.
- **Smart Participant Entry**: Manual name entry with auto-completion and quick-select tags.
- **CSRF Protection**: Comprehensive security for all booking actions.
- **V1-V2 Data Migration**: Seamless transition from legacy data structures.

---

## v2.0.1 - Stability & Housekeeping (In Progress)
*Focus: Performance, Cleanup, and UX Refinement*

### 1. ✅ Housekeeping (Legacy Code Removal) - COMPLETED
- [x] **Remove V1 Templates**: Deleted `templates/forms/bookTraining/` and associated assets.
- [x] **Cleanup Routes**: Removed legacy 3-layer architecture routers (`internal/handler`).
- [x] **Domain Simplification**: Stripped `internal/db/` and `internal/service/` legacy directories.
- [x] **Archive Migration**: Removed `v1tov2` utilities and decoupled Repository from migration logic.
- [x] **Architecture Alignment**: Migrated MCP and LINE messaging to the Transport layer.
- [x] **Utility Consolidation**: Unified `timeutil` and `lineutil` for shared cross-layer functions.

### 2. ✅ Performance Optimization - COMPLETED
- [x] **Parallel Fetching**: Integrated `errgroup` in `getBookingV2Form` to fetch stats, schedules, and bookings concurrently.
- [x] **Request Merging (SingleFlight)**: Implemented `SingleFlight` in both Handler and Repository layers to prevent cache stampedes and database spikes.
- [x] **Asynchronous Cache Invalidation**: Built a centralized `CacheWorker` pool (5 workers) for non-blocking background cleanup of user stats and schedules.
- [x] **Algorithmic Optimization**: Refactored `groupToWeeks` logic from $O(N \times M)$ to $O(N)$ using map lookups, reducing CPU time during scheduling.
- [x] **Concurrency Stability**: Secured global caches with `sync.RWMutex` and added `recover` blocks to background goroutines for auto-recovery.
- [x] **Observability Tuning**: Optimized OpenTelemetry sampling rate to 5% to reduce GC pressure and memory overhead at high throughput (4500+ RPS).

### 3. ✨ UX & Robustness (Next Focus)
- [ ] **Frontend Resource Separation**: Extract inline JavaScript from `booking_v2.templ` to `/assets/js/booking_v2.js`.
- [ ] **Toast System Integration**: Replace all `alert()` and `console.error()` calls with the project's native Toast notifications.
- [ ] **Skeleton Screens**: Add loading placeholders for smoother infinite scroll transitions.
- [ ] **Input Validation**: Implement stricter server-side validation for student names.

---

## v2.1.0+ - Future Considerations
- [ ] **Notification Center**: In-app notifications for booking approvals and changes.
- [ ] **Multi-language Support**: Full translation coverage for the V2 UI.
- [ ] **Admin Dashboard**: Enhanced management view for class capacities and attendance tracking.

---
*Roadmap last updated: 2026-02-22 by Gemini CLI*
