# Research: SSO Landing Page

## 1. Framer Motion Integration

### Findings

Framer Motion is the industry standard for React animations. For a landing page, "scroll-triggered" animations (reveal on scroll) are the primary use case.

### Best Practices

- **Performance**: Use `will-change` (handled by Framer Motion usually) and animate transform properties (`opacity`, `y`, `scale`) rather than layout properties (`height`, `width`).
- **Accessibility**: Respect `prefers-reduced-motion`. Framer Motion handles this with `useReducedMotion`.
- **Reusability**: Encapsulate animation logic in a wrapper component to keep page code clean.
- **Triggering**: Use `whileInView` prop for simple scroll triggers. Set `viewport={{ once: true, margin: "-100px" }}` to trigger slightly before the element hits the bottom of the viewport and only animate once.

### Code Snippets

**Reusable ScrollReveal Component:**

```tsx
import { motion, useInView } from "framer-motion";
import { useRef } from "react";

interface ScrollRevealProps {
  children: React.ReactNode;
  width?: "fit-content" | "100%";
  delay?: number;
}

export const ScrollReveal = ({
  children,
  width = "fit-content",
  delay = 0,
}: ScrollRevealProps) => {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: "-50px" });

  return (
    <div ref={ref} style={{ position: "relative", width, overflow: "hidden" }}>
      <motion.div
        variants={{
          hidden: { opacity: 0, y: 75 },
          visible: { opacity: 1, y: 0 },
        }}
        initial="hidden"
        animate={isInView ? "visible" : "hidden"}
        transition={{ duration: 0.5, delay: delay }}
      >
        {children}
      </motion.div>
    </div>
  );
};
```

**Alternative (Simpler `whileInView`):**

```tsx
<motion.div
  initial={{ opacity: 0, y: 20 }}
  whileInView={{ opacity: 1, y: 0 }}
  viewport={{ once: true }}
  transition={{ duration: 0.5 }}
>
  Content
</motion.div>
```

## 2. Smooth Scrolling Strategy

### Comparison

| Method                            | Pros                                                           | Cons                                                                                     |
| --------------------------------- | -------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| **CSS `scroll-behavior: smooth`** | Native, zero bundle size, easiest to implement.                | Less control over timing/easing. Offset for fixed headers requires `scroll-padding-top`. |
| **`react-scroll`**                | Precise control, active class handling (spy), offset handling. | Extra dependency (~3kb), imperative API.                                                 |
| **`react-router-hash-link`**      | Built for React Router, handles cross-page hash scrolling.     | Extra dependency.                                                                        |

### Recommendation

**Use Native CSS + React Router DOM**.
Since we are using Tailwind and want to keep dependencies low, native CSS is sufficient. We can handle the "fixed header offset" using Tailwind's `scroll-pt-*` utility on the `html` element.

### Implementation Details

1.  **Global CSS**: Add `html { scroll-behavior: smooth; }` (or Tailwind `scroll-smooth` class on `html`).
2.  **Header Offset**: Add `scroll-padding-top: [header-height]` to `html`. In Tailwind: `html { @apply scroll-smooth scroll-pt-16; }`.
3.  **Cross-page Hash Scrolling**: Since we have `/dashboard` and `/auth`, users might click a "Features" link from the login page. We need a small utility to handle scrolling to hash on mount if present.

```tsx
// src/components/ScrollToHashElement.tsx
import { useEffect } from "react";
import { useLocation } from "react-router-dom";

const ScrollToHashElement = () => {
  const { hash } = useLocation();

  useEffect(() => {
    if (hash) {
      const element = document.getElementById(hash.replace("#", ""));
      if (element) {
        element.scrollIntoView({ behavior: "smooth", block: "start" });
      }
    }
  }, [hash]);

  return null;
};
```

## 3. Routing Configuration

### Analysis

Currently, `App.tsx` redirects `/` to `/dashboard` or `/login`.
We need to:

1.  Remove the automatic redirect at `/`.
2.  Render `LandingPage` at `/`.
3.  `LandingPage` should handle the "Logged In" state UI (e.g., show "Go to Dashboard" button) rather than redirecting automatically. This is standard for SaaS landing pages (you can still see the landing page even if logged in).

### Recommended App.tsx Structure

```tsx
// ... imports

export function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <ScrollToHashElement /> {/* Handle hash scrolling */}
          <Routes>
            {/* Landing Page at root */}
            <Route path="/" element={<LandingPage />} />

            {/* Public routes */}
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/login" element={<LoginPage />} />
            {/* ... other auth routes */}

            {/* Protected routes */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardPage />
                </ProtectedRoute>
              }
            />

            {/* ... */}
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}
```

## Decisions

- **Animation**: Use **Framer Motion** with a reusable `ScrollReveal` wrapper component for section entrances. Use `whileInView` for simplicity.
- **Scrolling**: Use **Native CSS** (`scroll-smooth` + `scroll-pt-16`) for performance. Add a `ScrollToHashElement` component to handle deep links from other routes.
- **Routing**: Serve `LandingPage` at `/`. Remove `HomeRedirect`. The Landing Page will conditionally render "Dashboard" or "Login" buttons based on `isAuthenticated()`.
