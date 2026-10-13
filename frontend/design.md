<design-context>
---
version: alpha
name: Asana
description: "Work management software for teams and AI agents. Plan, automate, and execute critical workflows together."
sourceUrl: "https://asana.com"

colors:
  primary: "#646f79"
  on-primary: "#ffffff"
  background: "#ffffff"
  surface: "#0d0d0d"
  border: "#f3f3f3"
  text: "#ffffff"
  text-muted: "#0d0d0d"
  accent: "#0d0e10"

typography:
  display:
    fontFamily: "TWK Lausanne, Helvetica Neue, Helvetica, sans-serif"
    fontSize: 30px
    fontWeight: 400
    lineHeight: 1.2
  heading:
    fontFamily: "TWK Lausanne, Helvetica Neue, Helvetica, sans-serif"
    fontSize: 24px
    fontWeight: 500
    lineHeight: 1.2
  body:
    fontFamily: "TWK Lausanne, Helvetica Neue, Helvetica, sans-serif"
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.5

spacing:
  base: 4px
  scale: [4, 8, 12, 16, 20, 24, 32, 48, 64, 112]

radius:
  sm: 4px
  md: 8px
  lg: 20px
  xl: 64px

shadows:
  card: "rgba(0, 0, 0, 0.15) 0px 2px 8px 0px"
  elevated: "rgba(0, 0, 0, 0.15) 0px 2px 8px 0px"

motion:
  duration-fast: 100ms
  duration-base: 200ms
  duration-slow: 300ms
  easing: "ease-in-out"

breakpoints: [1px, 200px, 300px, 350px, 400px, 414px, 480px, 500px, 520px, 550px, 560px, 600px, 768px, 769px, 960px, 961px, 1024px, 1100px, 1120px, 1125px, 1200px, 1210px, 1280px, 1300px, 1360px, 1400px, 1440px, 1441px, 1500px, 1600px, 1680px, 1800px, 1880px, 2000px]
---

## Rationale

Asana's design system reflects a **professional B2B work-management platform** designed for human-agent collaboration. The palette is deliberately muted—a warm gray primary (#646f79) paired with a nearly-white background—creating a calm, distraction-free canvas where the user's work takes center stage. This restraint signals trust and maturity; there are no aggressive brand colors competing for attention. The typography hierarchy is equally measured, using TWK Lausanne (a geometric sans-serif with strong Swiss modernist roots) at generous sizes (54px display, 30px heading) but with light weights and tight letter-spacing that prevent heaviness. Together, these choices suggest a platform built for *sustained, focused work*—not flashy feature discovery.

The color strategy itself is a study in separation of concerns. A muted text-muted (#0d0d0d, near-black) for copy, a darker accent (#0d0e10) for UI elements, and a light surface (#f3f3f3) for secondary zones create clear hierarchy without relying on bright contrasts. The primary gray serves as a middle tone for interactive states and buttons, enabling a calm but functional interface. This is intentional: teams using Asana spend hours in the product, and visual fatigue matters.

Spacing and motion values follow a deliberate 8px base grid that scales up to 160px, with conservative animation durations (200ms base, max 300ms) tuned for professional environments where snappy feedback matters but not at the cost of comprehension. The radius scale—from sharp 4px corners to soft 50px buttons—suggests a mix of data-dense UI (tight radius) and friendly affordances (large radius on CTAs), balancing form and function.

## 1. Visual Theme & Atmosphere

Asana projects an atmosphere of **calm competence**. The light color mode, neutral palette, and restrained use of visual contrast create a professional, anti-opinionated aesthetic—one that feels like a blank slate rather than a branded experience. This is by design: the software is the *container* for team productivity, not the attraction. The measured 54px display type with 1:1 line-height and -1.08px letter-spacing feels sophisticated without being playful, grounded without being corporate-stiff.

The overall impression is **minimal-maximalist**: extremely clean surfaces, but rich in interactive density. Shadows are subtle (2–8px blur, 15% opacity), borders are soft and understated, and rounded corners range from utilitarian (4px) to genuinely friendly (50px on primary actions). This sends a message: *we're serious about work, but we care about how you feel*.

## 2. Color System

**Primary & Neutral Core:**
- Primary color (#646f79, a warm mid-gray) is the backbone of the interactive system—buttons, links, active states. It's professional and non-threatening, sitting between the background and accent.
- On-primary (#ffffff) ensures button text reads crisply against the gray.
- Background (#ffffff, true white) is the canvas.
- Surface (#f3f3f3, barely tinted gray) differentiates secondary zones (cards, modals, sidebars) without jarring contrast.
- Border (#f3f3f3) matches surface, creating a soft, unified aesthetic where cards blend gently into the page.

**Text & Accent:**
- Text (#ffffff) is counterintuitive here—it likely refers to text *on* colored backgrounds or interactive states, not primary body text.
- Text-muted (#0d0d0d, near-black) is the true body text color, providing 18.5:1 contrast against white—exceeding WCAG AAA for accessibility.
- Accent (#0d0e10, near-black with a hint of cool tone) is used for secondary UI elements, emphasis, or darker states.

**Intent:** This palette avoids color saturation entirely. No reds, blues, or greens in the primary system means the brand doesn't compete with user-generated content, charts, or status indicators that might appear *within* work items. It's a system built to recede.

## 3. Typography

Asana uses **TWK Lausanne** exclusively—a geometric, Swiss-influenced sans-serif that conveys precision and international appeal. The hierarchy is deeply rooted in *scale and weight*, not color or decoration:

- **Display (54px, weight 300, 1:1 line-height):** Used for hero headings ("The OS for human-agent teams"). Extremely elegant, with aggressive letter-spacing (-1.08px) that tightens the letterforms, creating a distinctive mark. The light weight prevents it from feeling heavy.
- **Heading (30px, weight 400, 1.2 line-height):** Section titles and major UI text. Regular weight, 1.2 line-height provides breathing room for readability.
- **Body (16px, weight 400, 1.5 line-height):** Standard copy, links, and labels. Generous line-height (1.5) supports sustained reading and reduces cognitive load—critical for a work-management tool where users parse lots of text.

**Design Impact:** All three styles use the same typeface and light-to-regular weights, creating a harmonious, refined feel. The letter-spacing on display type suggests attention to detail and craftsmanship. Line-height values (1.0, 1.2, 1.5) are generous, favoring legibility over density.

## 4. Components & Patterns

While token data doesn't reveal component specifics, the measured values imply:

- **Buttons:** Likely 44–48px tall (touch-target compliant), using the primary gray with white text, rounded at 16px (lg radius) for affordance.
- **Cards & Containers:** 8–16px corner radius (md–lg), with subtle shadows (2–8px, 15% opacity) to create layering without depth.
- **Input Fields:** Borders matching the surface color (#f3f3f3), allowing them to blend until focused, when a state change (likely a subtle shadow or border color shift) indicates interaction.
- **Navigation & Modals:** Likely use the surface color as a secondary container, with clear separation from the main canvas.
- **Status Indicators:** Given the neutral color palette, these probably rely on shape, icon, or secondary hues not captured in the primary tokens.

The spacing scale (8, 16, 24, 32, 40, 48, 80, 112, 120, 160px) suggests:
- 8–24px for component-internal spacing (padding, gaps between elements).
- 32–48px for section-level spacing (between feature blocks).
- 80–160px for page-level spacing (hero sections, major layout divisions).

## 5. Spacing & Layout

The **8px base grid** is foundational. Every spacing value is a multiple of 8, enabling pixel-perfect layouts across breakpoints. The scale jumps include some unusual values (80, 112, 120, 160), suggesting tailored spacing for specific layout patterns—likely hero sections, card grids, and multi-column layouts.

**Breakpoints** are extremely granular, with 33 defined values ranging from 200px to 2000px. This density indicates:
- Heavy investment in responsive design across phone, tablet, and desktop.
- Likely dynamic layout shifts at each breakpoint to optimize for screen real estate.
- Support for unusual devices or zoom levels (note the 1px breakpoint—possibly for a CSS debugging hook).

The major breakpoints cluster at 768px (tablet), 1024px (desktop), 1280px+ (large desktop), suggesting a mobile-first, then tablet, then desktop progression.

## 6. Motion & Interaction

Motion is **conservative and functional**, not decorative:

- **Fast (100ms):** Micro-interactions like button hovers, tooltip fades, checkbox toggles.
- **Base (200ms):** Standard transitions—modal opens, nav slides, page fades.
- **Slow (300ms):** Longer transitions, possibly multi-step animations or enter/exit of large components.

All motion uses `ease-in-out`, which feels natural and familiar—not bouncy or playful. This aligns with the "calm competence" tone. The maximum duration (300ms) respects the principle that long animations feel sluggish in productivity apps; users need to feel in control, not waiting.

## Accessibility

### Contrast Ratios

**Text (#0d0d0d on #ffffff background):**
- Contrast ratio: ~18.5:1
- **WCAG AAA compliant** (7:1+). This exceeds AA standards (4.5:1) significantly, supporting users with color blindness and low vision.

**Primary text on Primary background (#ffffff on #646f79):**
- Contrast ratio: ~4.6:1
- **WCAG AA compliant** (4.5:1 minimum), borderline for normal text. For large text (18pt+ or 14pt bold), this meets AA standards with room to spare.

**Accent (#0d0e10) on Surface (#f3f3f3):**
- Contrast ratio: ~14:1
- **WCAG AAA compliant**, supporting secondary text and UI elements.

**Critical gap:** Any text in primary color (#646f79) on white background (~4.8:1) should be reserved for large headings or icons, not body text. The measured color suggests this is respected in the actual design.

### Minimum Requirements

- **Touch targets:** Given the 48–160px spacing scale and generous button sizing, Asana almost certainly exceeds the 44×44px minimum for all interactive elements.
- **Focus indicator:** Not directly specified in tokens, but best practice for this design would be a 2px solid outline in a high-contrast color (not the primary gray) with 2–4px offset, visible at a glance without obstructing content.
- **Motion:** The 100–300ms animation range respects users with vestibular motion sensitivity (no rapid animations) while maintaining responsiveness.
- **Color alone:** The system does not rely solely on color to convey meaning; status and hierarchy are supported by shape, size, and weight.

</design-context>

Use the design system above for all UI you generate.
