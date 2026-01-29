# DDD Analysis Report

## Overview

The project currently exhibits a **Layered Architecture** with strong characteristics of the **Transaction Script** pattern. While there is a folder named `internal/domain`, the actual implementation does not follow Domain-Driven Design (DDD) principles. The business logic is primarily concentrated in the Service layer, operating directly on Database Models (Anemic Domain Model).

## Current State Analysis

### 1. Anemic Domain Model
- **Observation**: The structs in `internal/domain` (e.g., `Appointment`, `TrainingDate`) are largely data holders with little to no behavior.
- **Problem**: Logic regarding state changes (e.g., canceling an appointment, approving leave) is scattered across Service methods rather than encapsulated within the entities themselves.
- **Evidence**: `internal/domain/appointment.go` contains only struct definitions and `bson` tags. Logic like "check if user owns appointment" is found in `internal/service/training_date.go` (`AppointmentCancel` method).

### 2. Infrastructure Leakage into Domain
- **Observation**: Domain entities import `go.mongodb.org/mongo-driver/v2/bson`.
- **Problem**: The domain layer is coupled to specific persistence technology (MongoDB). This makes it hard to test in isolation or switch databases.
- **Evidence**: `internal/domain/appointment.go` imports `bson`.

### 3. Service Layer as Transaction Script
- **Observation**: The Service layer (`internal/service`) orchestrates everything. It calls the DB layer, manipulates data, and performs business checks.
- **Problem**: As complexity grows, services become bloated "God Classes" that are hard to maintain.
- **Evidence**: `TrainingDateService` handles everything from creating appointments, managing leaves, to formatting dates for display.

### 4. Database Model Usage
- **Observation**: The Service layer uses `internal/db/model` structs directly instead of `internal/domain` entities.
- **Problem**: This bypasses the domain layer entirely. Changes to the database schema (DB models) ripple directly into the business logic and potentially the API responses.
- **Evidence**: `TrainingDateService` methods return `*model.Appointment` and `*model.Aggr...` structs.

### 5. Aggregation/View Logic in Persistence Layer
- **Observation**: `internal/db/model` contains many `Aggr...` structs (e.g., `AggrTrainingDateHasAppoint`).
- **Problem**: These look like "View Models" or "Read Models" optimized for specific queries. Placing them in the DB model layer couples the database queries directly to the application's read needs.

## Recommendations for Refactoring to DDD

### Phase 1: Purify the Domain Layer
1.  **Remove Infrastructure Dependencies**: Strip `bson` tags and Mongo imports from `internal/domain`.
2.  **Define Value Objects**: Identify concepts like `TimeSlot` or `TrainingSession` that can be immutable value objects.

### Phase 2: Enrich Domain Models
1.  **Move Logic to Entities**:
    *   Move `Appointment.Cancel()` logic into the `Appointment` entity.
    *   Move `Leave.Create()` validation logic into the `Leave` entity (or a Factory).
2.  **Encapsulate State**: Make struct fields private where appropriate and expose methods to mutate state (e.g., `appointment.CheckIn()`).

### Phase 3: Implement Repository Pattern Correctly
1.  **Interface Definition**: Ensure Repository interfaces (in `domain`) accept and return **Domain Entities**, not DB Models.
2.  **Mapping**: Implement mappers (in `infrastructure/persistence`) to convert between Domain Entities and DB Models.

### Phase 4: Application/Service Layer Refinement
1.  **Orchestration Only**: The Service layer should only coordinate tasks (load entity -> call entity method -> save entity), not perform business logic.
2.  **DTOs**: Create specific DTOs for the API handlers to decouple the internal domain models from the external API contract.

## Proposed Directory Structure Changes

```text
internal/
├── domain/              <-- PURE DOMAIN (No DB tags, No framework code)
│   ├── appointment.go   <-- Entity with behavior
│   ├── training.go
│   └── repository.go    <-- Interfaces
├── application/         <-- SERVICE LAYER
│   ├── booking_service.go
│   └── dto/             <-- Data Transfer Objects
├── infrastructure/      <-- IMPLEMENTATION
│   ├── persistence/
│   │   ├── mongo/
│   │   │   ├── model/   <-- DB Models (with bson tags)
│   │   │   └── repo_impl.go
```
