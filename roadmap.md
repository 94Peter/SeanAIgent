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

### 2. ⚡ Performance Optimization (Next Focus)
- [ ] **Parallel Fetching**: Implement `errgroup` in `getBookingV2Form` to fetch stats, schedules, and bookings concurrently.
- [ ] **Frontend Resource Separation**: Extract inline JavaScript from `booking_v2.templ` to `/assets/js/booking_v2.js`.
- [ ] **API Caching**: Add short-term caching for monthly statistics.

### 3. ✨ UX & Robustness
- [ ] **Toast System Integration**: Replace all `alert()` and `console.error()` calls with the project's native Toast notifications.
- [ ] **Skeleton Screens**: Add loading placeholders for smoother infinite scroll transitions.
- [ ] **Input Validation**: Implement stricter server-side validation for student names.

---

## v2.1.0+ - Future Considerations
- [ ] **Notification Center**: In-app notifications for booking approvals and changes.
- [ ] **Multi-language Support**: Full translation coverage for the V2 UI.
- [ ] **Admin Dashboard**: Enhanced management view for class capacities and attendance tracking.

---
*Roadmap last updated: 2026-02-20 by Gemini CLI*
