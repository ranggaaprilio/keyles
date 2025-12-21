# Quickstart: SSO Landing Page

## Overview

The SSO Landing Page is the public entry point for the application. It is a single-page React component with smooth scrolling sections.

## Development

### Prerequisites

- Node.js 18+
- `npm install` to install dependencies (including new `framer-motion`)

### Running Locally

1. Start the development server:
   ```bash
   npm run dev
   ```
2. Open `http://localhost:5173/` (or configured port).
3. The Landing Page should be visible at the root path.

### Key Components

- **`src/pages/LandingPage.tsx`**: Main entry point.
- **`src/components/landing/`**: Contains all sections (Hero, About, Pricing, etc.).

### Adding a New Section

1. Create `NewSection.tsx` in `src/components/landing/`.
2. Add an `id` to the outer container (e.g., `id="new-section"`).
3. Import and add to `LandingPage.tsx`.
4. Add a link to the Navbar if needed.

### Updating Pricing

Edit the `tiers` array in `src/components/landing/Pricing.tsx`.

## Testing

Run unit tests:

```bash
npm run test
```
