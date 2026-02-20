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

## v2.0.1 - Stability & Housekeeping (Target)
*Focus: Performance, Cleanup, and UX Refinement*

### 1. 🧹 Housekeeping (Legacy Code Removal)
- [ ] **Remove V1 Templates**: Delete `templates/forms/bookTraining/` and associated assets.
- [ ] **Cleanup Routes**: Remove legacy `/training/booking` GET/POST handlers.
- [ ] **Domain Simplification**: Strip deprecated fields in `internal/booking/domain/entity` used only by V1.
- [ ] **Archive Migration**: Move `v1tov2` utilities to a dedicated `maintenance` package.

### 2. ⚡ Performance Optimization
- [ ] **Parallel Fetching**: Implement `errgroup` in `getBookingV2Form` to fetch stats, schedules, and bookings concurrently.
- [ ] **Frontend Resource Separation**: Extract inline JavaScript from `booking_v2.templ` to `/assets/js/booking_v2.js` for browser caching.
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
*Roadmap last updated: 2026-02-19 by Gemini CLI*
