# Tasks: SSO Landing Page

**Feature**: SSO Landing Page
**Status**: Draft
**Spec**: [spec.md](spec.md)

## Phase 1: Setup & Dependencies

- [x] T001 Install `framer-motion` dependency in `frontend/package.json`
- [x] T002 Initialize Shadcn UI in `frontend/` using `npx shadcn-ui@latest init`
- [x] T003 Install Shadcn components: `button`, `card`, `navigation-menu`, `separator`
- [x] T004 Create directory structure `frontend/src/components/landing/`

## Phase 2: Foundational Components

- [x] T005 [P] Create `ScrollReveal` component in `frontend/src/components/landing/ScrollReveal.tsx` using Framer Motion
- [x] T006 [P] Create `ScrollToHashElement` utility in `frontend/src/components/landing/ScrollToHashElement.tsx`
- [x] T007 [P] Create `Footer` component in `frontend/src/components/landing/Footer.tsx` using Shadcn `Separator`
- [x] T008 [P] Create `Navbar` component shell in `frontend/src/components/landing/Navbar.tsx` using Shadcn `NavigationMenu` (links only)

## Phase 3: User Story 1 - Learn About SSO Functionality (P1)

**Goal**: Users can understand SSO benefits and workflow via Hero and Explanation sections.
**Independent Test**: Verify Landing Page renders with Hero and Explanation sections containing static visual aids.

- [x] T009 [US1] Create `Hero` component in `frontend/src/components/landing/Hero.tsx` with headline and subheadline
- [x] T010 [US1] Create `SSOExplanation` component in `frontend/src/components/landing/SSOExplanation.tsx` with Lucide icons and 3-step flow
- [x] T011 [US1] Create `LandingPage` container in `frontend/src/pages/LandingPage.tsx` assembling Navbar, Hero, Explanation, Footer
- [x] T012 [US1] Add `ScrollReveal` animations to `Hero` and `SSOExplanation` components

## Phase 4: User Story 2 - Access Login (P1)

**Goal**: Existing users can access login; new users can access registration.
**Independent Test**: Verify "Login" button goes to `/login` (or Dashboard if auth) and "Get Started" goes to `/register`.

- [x] T013 [US2] Update `Navbar` in `frontend/src/components/landing/Navbar.tsx` to include "Login" button with auth state logic
- [x] T014 [US2] Update `Hero` in `frontend/src/components/landing/Hero.tsx` to include "Get Started" CTA with auth state logic
- [x] T015 [US2] Update `App.tsx` to serve `LandingPage` at root `/` and remove automatic redirect
- [x] T016 [US2] Add `ScrollToHashElement` to `App.tsx` to enable hash linking

## Phase 5: User Story 3 - Learn About the Service (P2)

**Goal**: Users can read about the company and mission.
**Independent Test**: Verify "About" section is visible and accessible via Navbar link.

- [x] T017 [US3] Create `About` component in `frontend/src/components/landing/About.tsx` using Shadcn `Card`
- [x] T018 [US3] Add `About` section to `frontend/src/pages/LandingPage.tsx` with ID `about`
- [x] T019 [US3] Update `Navbar` in `frontend/src/components/landing/Navbar.tsx` with smooth scroll link to `#about`

## Phase 6: User Story 4 - Review Pricing Options (P2)

**Goal**: Users can view and compare pricing tiers.
**Independent Test**: Verify "Pricing" section shows 3 tiers and is accessible via Navbar link.

- [x] T020 [US4] Create `Pricing` component in `frontend/src/components/landing/Pricing.tsx` using Shadcn `Card` for 3 tiers (Enterprise tier MUST use "Contact Sales" CTA)
- [x] T021 [US4] Add `Pricing` section to `frontend/src/pages/LandingPage.tsx` with ID `pricing`
- [x] T022 [US4] Update `Navbar` in `frontend/src/components/landing/Navbar.tsx` with smooth scroll link to `#pricing`

## Phase 7: Polish & Cross-Cutting

- [x] T023 Add global CSS for `scroll-behavior: smooth` in `frontend/src/index.css` (or via Tailwind config)
- [x] T024 Verify responsive layout for all sections on mobile and tablet
- [x] T025 Audit accessibility (ARIA labels, keyboard navigation for scroll links)

## Dependencies

- **US1** depends on **Foundational**
- **US2** depends on **US1** (needs Hero/Navbar to exist)
- **US3** depends on **US1** (needs LandingPage)
- **US4** depends on **US1** (needs LandingPage)

## Parallel Execution Examples

- **T005, T006, T007** can be implemented in parallel (independent components).
- **T017 (About)** and **T020 (Pricing)** can be implemented in parallel after US1 is complete.

## Implementation Strategy

1.  **Setup**: Get the environment ready with Shadcn and Framer Motion.
2.  **MVP (US1 + US2)**: Build the core page with Hero, Explanation, and Auth routing. This delivers the primary value.
3.  **Expansion (US3 + US4)**: Add About and Pricing sections to complete the content.
4.  **Polish**: Finalize animations and responsiveness.
