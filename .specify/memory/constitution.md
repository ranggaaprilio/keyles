<!--
Sync Impact Report - Constitution Update
=========================================
Version Change: 1.0.0 → 1.0.1
Modified Principles:
  - I. Clean Architecture: distinguish domain rules from use-case orchestration
  - V. Frontend Code Conventions: permit PascalCase route screens in src/pages/
  - Technology Stack Requirements: align contracts and API logic paths with repository layout
Added Sections: None
Removed Sections: None
Templates Status:
  - plan-template.md - Compatible with clarified layer responsibilities
  - spec-template.md - Compatible with current structure
  - tasks-template.md - Updated to require constitution-mandated tests
  - agent-file-template.md - No change required
  - checklist-template.md - No change required
Follow-up Items: None
-->

# Keyles Constitution

## Core Principles

### I. Clean Architecture (NON-NEGOTIABLE)

Backend structure MUST follow Clean Architecture or Hexagonal Architecture principles:

- **Domain/Entities** (innermost layer) MUST be completely independent of frameworks, databases, and UI
- **Domain rules and entities** MUST reside in the Domain layer with zero external dependencies
- **Application-specific orchestration** MAY reside in the Usecase layer and MUST depend
  only on Domain abstractions for outbound operations
- **Infrastructure concerns** (database, external APIs, UI) MUST remain in outer layers
- **Dependency direction** MUST always point inward toward the Domain layer
- Outer layers (Handlers, Controllers) MAY depend on Domain, but Domain MUST NEVER depend on outer layers

**Rationale**: Ensures business logic portability, testability, and long-term maintainability regardless of technology changes.

### II. SOLID Principles (NON-NEGOTIABLE)

All code in both Golang and React/TypeScript stacks MUST adhere to SOLID principles:

- **Single Responsibility Principle**: Each module/class has one reason to change
- **Open/Closed Principle**: Open for extension, closed for modification
- **Liskov Substitution Principle**: Subtypes must be substitutable for base types
- **Interface Segregation Principle**: Clients MUST NOT depend on interfaces they do not use
- **Dependency Inversion Principle (DIP)**: Domain MUST depend on abstractions, NOT concrete implementations

**Critical for DIP**:

- Domain layer MUST use interfaces for all outbound dependencies (repositories, external services)
- Infrastructure layer provides concrete implementations
- NEVER call database or external service implementations directly from domain logic

**Rationale**: SOLID principles ensure maintainable, scalable, and testable code architecture across the entire stack.

### III. Dependency Inversion via Interfaces (NON-NEGOTIABLE)

**Golang Backend**:

- All outbound dependencies (database repositories, external services) MUST be defined as interfaces in the Domain layer
- Concrete implementations MUST be in Infrastructure layer
- Domain functions MUST accept interface parameters, NEVER concrete types

**React/TypeScript Frontend**:

- Use abstractions/interfaces for all external dependencies (API clients, storage services)
- Components MUST depend on abstract contracts, not concrete implementations

**Rationale**: Enables independent testing of business logic and decouples domain from infrastructure details.

### IV. Backend Code Conventions (Golang)

**Mandatory Standards**:

- Follow [Effective Go](https://go.dev/doc/effective_go) and official Go best practices
- Package names MUST be lowercase, single-word when possible
- All exported functions, types, and methods MUST have documentation comments
- Use standard Go project layout for Clean Architecture:
  ```
  backend/
  ├── domain/          # Entities, business logic, interfaces
  ├── usecase/         # Application business rules
  ├── infrastructure/  # DB, external APIs, concrete implementations
  └── interfaces/      # HTTP handlers, controllers
  ```

**Rationale**: Consistency with Go community standards ensures code readability and maintainability.

### V. Frontend Code Conventions (React/TypeScript)

**Mandatory Standards**:

- Use TypeScript exclusively (no plain JavaScript)
- Use functional components and React Hooks exclusively
- All components MUST use PascalCase naming
- Reusable component files MUST be placed in the `src/components/` directory structure
- Route-level screen files MAY be placed in `src/pages/`; all component and screen files
  MUST use PascalCase naming
- Prefer composition over inheritance
- Use proper TypeScript typing (avoid `any` unless absolutely necessary)

**Rationale**: Modern React patterns with TypeScript provide type safety and improved developer experience.

### VI. Test Coverage Requirements (NON-NEGOTIABLE)

**Unit Testing**:

- ALL business logic MUST have unit tests
- Minimum 85% code coverage for Domain/Business Logic layer
- Tests MUST be independent and reproducible

**Integration Testing**:

- Every backend handler/controller MUST have integration tests
- Integration tests MUST verify infrastructure layer connections work correctly
- Frontend API integration points MUST be tested

**Testing Tools**:

- **Backend**: Go's built-in `testing` package or GoMock for mocking
- **Frontend**: Vitest and React Testing Library

**Test-First Workflow**:

- Write tests before implementation where feasible
- Ensure tests fail before writing the implementation
- Follow Red-Green-Refactor cycle

**Rationale**: High test coverage protects business logic integrity and prevents regressions during refactoring.

### VII. Architecture Validation Gates

Every feature plan MUST pass these validation gates before implementation:

**Clean Architecture Compliance**:

- [ ] Domain layer has no imports from infrastructure or frameworks
- [ ] All repository/service interfaces defined in Domain
- [ ] Concrete implementations only in Infrastructure layer
- [ ] Dependency arrows point inward

**SOLID Compliance**:

- [ ] Each module has single, well-defined responsibility
- [ ] Domain depends only on abstractions (interfaces)
- [ ] No direct database/external API calls from business logic

**Testing Requirements**:

- [ ] Unit test plan for all business logic (target: 85%+ coverage)
- [ ] Integration test plan for all handlers/controllers
- [ ] Test isolation strategy documented

**Rationale**: Proactive validation prevents architectural drift and ensures principles are enforced from design phase.

## Technology Stack Requirements

**Backend**:

- **Language**: Golang (latest stable version)
- **Architecture**: Clean Architecture / Hexagonal Architecture
- **Package Structure**: Standard Go layout with domain/usecase/infrastructure separation
- **Testing**: Go testing package + GoMock
- **Documentation**: Follow Effective Go conventions

**Frontend**:

- **Language**: TypeScript (strict mode enabled)
- **Framework**: React with Functional Components and Hooks
- **Testing**: Vitest + React Testing Library
- **Code Organization**: Reusable components in `src/components/`; route-level screens
  MAY be organized in `src/pages/`

**Database**:

- Repository pattern MUST be used
- Database access MUST be abstracted behind interfaces defined in Domain layer
- Infrastructure layer provides concrete repository implementations

**API Design**:

- RESTful or GraphQL feature contracts defined in `specs/<feature>/contracts/`
- API layer serves as boundary between frontend and backend
- Handlers/Controllers in outer layer, application orchestration in Usecase layer, and
  dependency-free domain rules in Domain layer

## Quality Gates

**Pre-Implementation Gates** (from Constitution Check in plan.md):

- Clean Architecture validation completed
- SOLID principles verified in design
- Interface definitions complete for all external dependencies
- Test plan approved with coverage targets

**Implementation Gates**:

- Code review MUST verify:
  - No domain-to-infrastructure dependencies
  - All outbound calls use interfaces
  - Proper layer separation maintained
  - Naming follows language conventions (Go: lowercase packages, exported docs; React: PascalCase components)

**Testing Gates**:

- Unit tests pass with ≥85% domain layer coverage
- Integration tests verify infrastructure connections
- No skipped or disabled tests without documented justification

**Deployment Gates**:

- All tests passing
- Code coverage requirements met
- Architecture validation confirmed
- Documentation complete for exported functions/components

## Governance

**Constitution Authority**:

- This constitution supersedes all other development practices and guidelines
- All feature specifications, plans, and implementations MUST comply with these principles
- Constitution violations require explicit justification and approval before proceeding

**Amendment Process**:

- Amendments require documentation of rationale and impact analysis
- Version number MUST increment per semantic versioning (MAJOR.MINOR.PATCH)
- All affected templates and documentation MUST be updated for consistency
- Migration plan required for breaking changes

**Compliance Enforcement**:

- All pull requests MUST include Constitution Check verification
- Code reviewers MUST validate architectural principles compliance
- Complexity that appears to violate principles MUST be justified in Complexity Tracking section of plan.md
- Regular architecture reviews to prevent drift

**Development Workflow**:

- Feature specs undergo Constitution Check before planning
- Implementation plans validate Clean Architecture and SOLID compliance
- Test-first approach enforced through task ordering
- Integration and unit tests required before feature completion

**Version**: 1.0.1 | **Ratified**: 2025-11-30 | **Last Amended**: 2026-06-01
