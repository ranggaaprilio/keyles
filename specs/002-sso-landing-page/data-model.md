# Data Model: SSO Landing Page

**Feature**: SSO Landing Page
**Type**: UI Component Props & Interfaces

## Component Interfaces

### 1. Pricing Tier

Used in `Pricing.tsx` to render pricing cards.

```typescript
interface PricingFeature {
  text: string;
  included: boolean;
}

interface PricingTier {
  name: string;
  price: string;
  description: string;
  features: PricingFeature[];
  ctaText: string;
  popular?: boolean; // Highlights the card
}
```

### 2. Feature Item

Used in `SSOExplanation.tsx` or `Hero.tsx` for feature lists.

```typescript
interface FeatureItem {
  title: string;
  description: string;
  icon: React.ElementType; // Lucide icon component
}
```

### 3. Navbar Props

Used in `Navbar.tsx`.

```typescript
interface NavbarProps {
  transparent?: boolean; // If true, background is transparent until scroll
}
```

### 4. ScrollReveal Props

Used in `ScrollReveal.tsx` wrapper.

```typescript
interface ScrollRevealProps {
  children: React.ReactNode;
  width?: "fit-content" | "100%";
  delay?: number; // Animation delay in seconds
}
```

## State Management

### Auth Store (Existing)

The landing page will read from the existing `useAuthStore` to determine if the user is logged in.

```typescript
// Existing store usage
const { isAuthenticated, user } = useAuthStore();
```

- **If `isAuthenticated` is true**:
  - "Login" button in Navbar becomes "Dashboard" (links to `/dashboard`).
  - "Get Started" button in Hero becomes "Go to Dashboard".
- **If `isAuthenticated` is false**:
  - "Login" button links to `/login`.
  - "Get Started" button links to `/register`.
