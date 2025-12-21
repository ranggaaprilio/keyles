# Implementation Plan: SSO Landing Page

**Branch**: `002-sso-landing-page` | **Date**: December 21, 2025 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-sso-landing-page/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Create a public-facing SSO Landing Page to educate users about the service, provide pricing information, and drive conversions. The page will be a single scrollable view with smooth navigation, built using React, Shadcn UI (Tailwind CSS), and Framer Motion for animations. It includes Hero, SSO Explanation, About, and Pricing sections, with clear CTAs for Login (existing users) and Registration (new users).

## Technical Context

**Language/Version**: TypeScript 5.x, React 18.x
**Primary Dependencies**:

- `react`, `react-dom` (Core)
- `tailwindcss` (Styling)
- `shadcn-ui` (UI Components - NEW)
- `lucide-react` (Icons)
- `react-router-dom` (Routing)
- `framer-motion` (Animation - NEW)
  **Storage**: N/A (Static content)
  **Testing**: `vitest`, `@testing-library/react`
  **Target Platform**: Web (Responsive: Mobile, Tablet, Desktop)
  **Project Type**: Web Frontend
  **Performance Goals**: Core content load < 2s on 3G, 95+ Accessibility Score
  **Constraints**: Single page layout, smooth scrolling, static assets only

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

**Clean Architecture Compliance**:

- [x] Domain layer has no imports from infrastructure or frameworks (N/A - Frontend UI feature)
- [x] All repository/service interfaces defined in Domain layer (N/A)
- [x] Concrete implementations only in Infrastructure layer (N/A)
- [x] Dependency arrows verified to point inward (toward Domain) (N/A)

**SOLID Principles Compliance**:

- [x] Each module has single, well-defined responsibility (SRP) - Components separated by section
- [x] Domain depends only on abstractions/interfaces (DIP) - N/A
- [x] No direct database/external API calls from business logic - N/A
- [x] Interface segregation verified for all contracts - Component props defined specifically

**Testing Requirements**:

- [x] Unit test plan documented for all business logic (target: ≥85% coverage) - UI tests for components
- [x] Integration test plan for all handlers/controllers - N/A
- [x] Test isolation strategy defined (mocking approach) - N/A
- [x] Test-first workflow feasible for this feature - Yes

**Code Conventions**:

- [x] Backend: Follows Effective Go, lowercase packages, exported function docs (N/A)
- [x] Frontend: TypeScript strict mode, PascalCase components, functional components only
- [x] Clear separation between backend (domain/usecase/infrastructure) and frontend (components/services)

**Violations Requiring Justification** (fill Complexity Tracking section if any):

- [x] No constitution violations OR all violations documented with justification

## Project Structure

### Documentation (this feature)

```text
specs/002-sso-landing-page/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (UI Component Props)
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
frontend/
├── src/
│   ├── components/
│   │   └── landing/     # New directory for landing page components
│   │       ├── Hero.tsx
│   │       ├── SSOExplanation.tsx
│   │       ├── About.tsx
│   │       ├── Pricing.tsx
│   │       ├── Footer.tsx
│   │       └── Navbar.tsx
│   ├── pages/
│   │   └── LandingPage.tsx # Main page container
│   └── App.tsx          # Routing updates
```

## Phase 0: Research & Validation

1. [ ] **Research Framer Motion**: Determine best practices for scroll-triggered animations (fade-in, slide-up) compatible with Tailwind CSS.
2. [ ] **Research Smooth Scrolling**: Identify best method for smooth scrolling to anchor tags in React (native CSS `scroll-behavior: smooth` vs. React libraries).
3. [ ] **Routing Strategy**: Verify how to serve `LandingPage` at root `/` while preserving access to `/dashboard` and `/auth/*` routes in `App.tsx`.
4. [ ] **Shadcn UI Setup**: Verify installation steps for Shadcn UI in the existing Vite project (init command, components.json).

## Phase 1: Design & Contracts

1. **Data Model**: Define component props (e.g., `PricingTier` interface, `FeatureItem` interface) in `data-model.md`.
2. **Assets**: List required Lucide icons for SSO explanation diagrams.
3. **Agent Context**: Update context to include `framer-motion` and `shadcn-ui`.

## Phase 2: Implementation

### Step 1: Setup & Dependencies

- Install `framer-motion`.
- Initialize Shadcn UI: `npx shadcn-ui@latest init`.
- Install required Shadcn components: `button`, `card`, `navigation-menu`, `separator`.
- Create `src/components/landing/` directory.

### Step 2: Component Development

- **Navbar**: Sticky header using Shadcn `NavigationMenu` with "About", "Pricing" links (scroll) and "Login" `Button` (route).
- **Hero**: Headline, subheadline, "Get Started" Shadcn `Button` CTA, animated entrance.
- **SSO Explanation**: 3-step visual flow using Lucide icons and CSS shapes.
- **About**: Text content and company info using Shadcn `Card` for layout if applicable.
- **Pricing**: 3 Shadcn `Card` components (Free, Pro, Enterprise) with feature lists and `Button` CTAs.
- **Footer**: Copyright and links using Shadcn `Separator` for layout.

### Step 3: Page Assembly

- Create `LandingPage.tsx` composing all sections.
- Implement smooth scrolling behavior.

### Step 4: Routing & Integration

- Update `App.tsx` to route `/` to `LandingPage`.
- Ensure protected routes remain secure.

### Step 5: Polish

- Add `framer-motion` scroll animations to sections.
- Verify responsive layout on mobile/tablet.
- Audit accessibility (ARIA labels, keyboard nav).

## Phase 3: Documentation

- Update `README.md` with new dependency info.
  ├── usecase/ # Application business rules
  ├── infrastructure/ # Concrete implementations (DB, external APIs)
  │ ├── persistence/ # Repository implementations
  │ └── external/ # External service clients
  ├── interfaces/ # Handlers, controllers (outer layer)
  │ └── http/
  └── tests/
  ├── unit/
  │ ├── domain/
  │ └── usecase/
  ├── integration/
  └── contract/

frontend/
├── src/
│ ├── components/ # PascalCase React components
│ ├── services/ # API client abstractions
│ ├── types/ # TypeScript type definitions
│ └── hooks/ # Custom React hooks
└── tests/
├── unit/
└── integration/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)

api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]

```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
```
