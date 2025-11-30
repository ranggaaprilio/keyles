---

description: "Task list template for feature implementation"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[###-feature-name]/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Per constitution, tests are MANDATORY. All business logic requires unit tests (≥85% coverage), and all handlers require integration tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `src/`, `tests/` at repository root
- **Web app (Clean Architecture)**: 
  - Backend: `backend/domain/`, `backend/usecase/`, `backend/infrastructure/`, `backend/interfaces/`
  - Frontend: `frontend/src/components/`, `frontend/src/services/`
  - Tests: `backend/tests/unit/`, `backend/tests/integration/`, `frontend/tests/`
- **Mobile**: `api/src/`, `ios/src/` or `android/src/`
- Paths shown below assume web app with Clean Architecture - adjust based on plan.md structure

<!-- 
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.
  
  The /speckit.tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements from plan.md
  - Entities from data-model.md
  - Endpoints from contracts/
  
  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Tested independently
  - Delivered as an MVP increment
  
  DO NOT keep these sample tasks in the generated tasks.md file.
  ============================================================================
-->

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project structure per implementation plan
- [ ] T002 Initialize [language] project with [framework] dependencies
- [ ] T003 [P] Configure linting and formatting tools

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

Examples of foundational tasks (adjust based on your project):

- [ ] T004 Setup database schema and migrations framework
- [ ] T005 [P] Define repository interfaces in Domain layer (Clean Architecture)
- [ ] T006 [P] Setup API routing and middleware structure
- [ ] T007 Create base domain entities and business logic interfaces
- [ ] T008 Implement repository concrete implementations in Infrastructure layer
- [ ] T009 Configure error handling and logging infrastructure
- [ ] T010 Setup environment configuration management

**Clean Architecture Checkpoint**: 
- Domain layer established with interfaces only
- Infrastructure layer provides implementations
- No domain-to-infrastructure dependencies

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 1 (MANDATORY per constitution) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**
> **Target: ≥85% coverage for domain/business logic, integration tests for all handlers**

- [ ] T010 [P] [US1] Unit tests for domain entities in tests/unit/domain/test_[entity]_test.go
- [ ] T011 [P] [US1] Unit tests for use cases/business logic in tests/unit/usecase/test_[usecase]_test.go
- [ ] T012 [P] [US1] Integration test for handler in tests/integration/test_[handler]_test.go
- [ ] T013 [P] [US1] Contract test for API endpoint in tests/contract/test_[endpoint]_test.go

### Implementation for User Story 1

> **Clean Architecture Task Order**: Domain → Use Cases → Infrastructure → Interfaces (Handlers)

- [ ] T014 [P] [US1] Define domain entities in backend/domain/entities/[entity].go
- [ ] T015 [P] [US1] Define repository interfaces in backend/domain/repositories/[repository].go
- [ ] T016 [US1] Implement use case/business logic in backend/usecase/[usecase].go (depends on T014, T015)
- [ ] T017 [US1] Implement repository concrete class in backend/infrastructure/persistence/[repository]_impl.go
- [ ] T018 [US1] Implement handler in backend/interfaces/http/[handler].go (outer layer)
- [ ] T019 [US1] Add validation and error handling in domain layer
- [ ] T020 [US1] Add logging for user story 1 operations (infrastructure concern)
- [ ] T021 [P] [US1] Frontend: Create TypeScript components in frontend/src/components/[Component].tsx
- [ ] T022 [US1] Frontend: Create service abstraction for API calls in frontend/src/services/[service].ts

**Architecture Verification**:
- Domain has no infrastructure imports ✓
- Use cases depend only on domain interfaces ✓
- Handlers depend on use cases, not domain directly ✓

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 2 (MANDATORY per constitution) ⚠️

- [ ] T023 [P] [US2] Unit tests for domain entities in tests/unit/domain/test_[entity]_test.go
- [ ] T024 [P] [US2] Unit tests for use cases in tests/unit/usecase/test_[usecase]_test.go
- [ ] T025 [P] [US2] Integration test for handler in tests/integration/test_[handler]_test.go

### Implementation for User Story 2

- [ ] T026 [P] [US2] Define domain entities in backend/domain/entities/[entity].go
- [ ] T027 [P] [US2] Define interfaces in backend/domain/repositories/[repository].go
- [ ] T028 [US2] Implement use case in backend/usecase/[usecase].go
- [ ] T029 [US2] Implement infrastructure in backend/infrastructure/[implementation].go
- [ ] T030 [US2] Implement handler in backend/interfaces/http/[handler].go
- [ ] T031 [US2] Integrate with User Story 1 components (if needed, via interfaces only)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - [Title] (Priority: P3)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 3 (MANDATORY per constitution) ⚠️

- [ ] T032 [P] [US3] Unit tests for domain in tests/unit/domain/test_[entity]_test.go
- [ ] T033 [P] [US3] Unit tests for use cases in tests/unit/usecase/test_[usecase]_test.go
- [ ] T034 [P] [US3] Integration test for handler in tests/integration/test_[handler]_test.go

### Implementation for User Story 3

- [ ] T035 [P] [US3] Define domain entities in backend/domain/entities/[entity].go
- [ ] T036 [P] [US3] Define interfaces in backend/domain/[interfaces].go
- [ ] T037 [US3] Implement use case in backend/usecase/[usecase].go
- [ ] T038 [US3] Implement infrastructure in backend/infrastructure/[implementation].go
- [ ] T039 [US3] Implement handler in backend/interfaces/http/[handler].go

**Checkpoint**: All user stories should now be independently functional

---

[Add more user story phases as needed, following the same pattern]

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] TXXX [P] Documentation updates in docs/ (Effective Go compliance, component docs)
- [ ] TXXX Code cleanup and refactoring (maintain Clean Architecture)
- [ ] TXXX Performance optimization across all stories
- [ ] TXXX [P] Verify ≥85% test coverage for domain layer
- [ ] TXXX Security hardening
- [ ] TXXX Architecture compliance review (no domain-to-infrastructure dependencies)
- [ ] TXXX Run quickstart.md validation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (if tests requested):
Task: "Contract test for [endpoint] in tests/contract/test_[name].py"
Task: "Integration test for [user journey] in tests/integration/test_[name].py"

# Launch all models for User Story 1 together:
Task: "Create [Entity1] model in src/models/[entity1].py"
Task: "Create [Entity2] model in src/models/[entity2].py"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
