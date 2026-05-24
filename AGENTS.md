# Repository Guidelines

## Project Structure & Module Organization

Keyles is a multi-tenant SSO platform with separate backend, frontend, and specification areas. `backend/` contains the Go service, organized by Clean Architecture: `cmd/` entry points, `domain/` entities and interfaces, `usecase/` application logic, `infrastructure/` external implementations, `interfaces/http/` handlers and routing, `migrations/` SQL migrations, and `tests/integration/`. `frontend/` contains the React + TypeScript Vite app under `frontend/src/`, including `components/`, `hooks/`, `services/`, `stores/`, `types/`, and `tests/`. Product specs and API contracts are under `specs/`.

## Build, Test, and Development Commands

From `backend/`:

- `make run` starts the API server; `make build` builds `bin/server`.
- `make test` runs Go tests; `make test-coverage` writes coverage reports.
- `make docker-up` / `make docker-down` manage local Docker services.
- `make migrate-up`, `make migrate-down`, and `make seed` manage database state.

From `frontend/`:

- `npm run dev` starts Vite on port 5173.
- `npm run build` type-checks and builds production assets.
- `npm run lint` runs ESLint with zero warnings allowed.
- `npm run test` runs Vitest in `jsdom`.

## Coding Style & Naming Conventions

Format Go code with `gofmt`; keep packages lowercase and focused by layer. Preserve dependency direction: domain code must not import infrastructure or HTTP packages. Name migrations like `000013_describe_change.up.sql` and matching `.down.sql`.

Frontend code uses TypeScript, React function components, hooks, Tailwind CSS, and the `@/` alias. Name components in PascalCase (`ClientManagement.tsx`), hooks as `useThing.ts`, and services by domain (`user.ts`, `oauthService.ts`). Prefer existing Radix UI and `components/ui/` primitives.

## Testing Guidelines

Backend tests use Go’s standard test runner. Put integration tests in `backend/tests/integration/` and name files `*_test.go`. Run `make test` before submitting; use `make test-coverage` for broader changes.

Frontend tests use Vitest with Testing Library setup in `frontend/src/tests/setup.ts`. Keep tests close to user-visible behavior and run `npm run test` plus `npm run lint`.

## Commit & Pull Request Guidelines

History uses short imperative commit subjects, often with Conventional Commit prefixes such as `feat:` and `refactor:`. Keep subjects concise and scoped, for example `feat: add client secret rotation`.

Pull requests should include a brief summary, linked issue or spec when relevant, test results, and screenshots for UI changes. Note migrations, new environment variables, or security-sensitive changes explicitly.

## Security & Configuration Tips

Never commit `.env` files, generated private keys, or secrets. Backend RSA keys are generated with `make keygen` and should stay under ignored local paths such as `backend/keys/`. Use `.env.example` files as templates.
