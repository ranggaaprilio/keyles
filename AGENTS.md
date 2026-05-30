# Repository Guidelines

## Project Structure & Module Organization

Keyles is a multi-tenant SSO platform with a Go backend and Vite/React frontend. Backend code lives in `backend/`: `cmd/` contains entry points, `domain/` defines entities and interfaces, `usecase/` holds application rules, `infrastructure/` contains adapters, `interfaces/http/` exposes routes, `migrations/` stores SQL migrations, and `tests/` contains unit, integration, and mock packages. Frontend code lives in `frontend/src/`, organized by `components/`, `pages/`, `hooks/`, `services/`, `stores/`, `types/`, and `utils/`; email templates are in `frontend/emails/`. Product specs, plans, and API contracts are under `specs/`.

## Build, Test, and Development Commands

Use Docker Compose from the repository root for the full stack:

```bash
docker-compose up -d
docker-compose logs -f backend
docker-compose down
```

Backend commands run from `backend/`: `make run` starts the API, `make build` writes `bin/server`, `make test` runs `go test -v ./...`, `make test-coverage` generates `coverage.out` and `coverage.html`, and `make migrate-up` applies SQL migrations. Frontend commands run from `frontend/`: `npm run dev` starts Vite on port 5173, `npm run build` type-checks and builds, `npm run lint` runs ESLint with zero warnings allowed, `npm test` runs Vitest, and `npm run preview` serves the production build.

## Coding Style & Naming Conventions

Format Go with `gofmt`; keep packages lowercase and focused by layer. Keep domain interfaces free of framework or database dependencies. TypeScript uses React functional components, PascalCase component/page filenames, camelCase hooks and helpers, and the `@/` alias for `frontend/src`. Match existing two-space TSX formatting and single-quote style. Prefer shared UI primitives in `frontend/src/components/ui/` before adding new component patterns.

## Testing Guidelines

Backend tests use Go’s testing package plus `testify`; name files `*_test.go` and place broader API scenarios in `backend/tests/integration/`. Run `make test` before backend changes and `make test-coverage` for risky usecase or domain work. Frontend tests use Vitest with jsdom and Testing Library setup in `frontend/src/tests/setup.ts`; use `*.test.tsx` or `*.test.ts` near the code or under `src/tests/`.

## Commit & Pull Request Guidelines

History follows concise conventional-style subjects such as `feat:`, `fix:`, `refactor:`, and scoped feature IDs like `feat(005):`. Keep commits imperative and focused. PRs should describe the change, link related spec or issue IDs, list test commands run, and include screenshots or short recordings for user-facing UI changes. Note any migration, environment, or Docker Compose changes explicitly.

## Security & Configuration Tips

Never commit real `.env` values or API keys. Start from `backend/.env.example` and `frontend/.env.example`; keep secrets in local environment files or deployment configuration.

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
<!-- SPECKIT END -->
