/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    './pages/**/*.{ts,tsx}',
    './components/**/*.{ts,tsx}',
    './app/**/*.{ts,tsx}',
    './src/**/*.{ts,tsx}',
  ],
  prefix: "",
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        // Dell 1996 brand colors
        "dell-red": "hsl(var(--dell-red))",
        "dell-yellow": "hsl(var(--dell-yellow))",
        "dell-purple": "hsl(var(--dell-purple))",

        // Surface
        "frame-ink": "hsl(var(--frame-ink))",
        "canvas": "hsl(var(--canvas))",

        // Text
        ink: "hsl(var(--ink))",
        link: "hsl(var(--link))",

        // Ribbon-card tint family
        "tint-olive": "hsl(var(--tint-olive))",
        "tint-sage": "hsl(var(--tint-sage))",
        "tint-salmon": "hsl(var(--tint-salmon))",
        "tint-peach": "hsl(var(--tint-peach))",
        "tint-lime": "hsl(var(--tint-lime))",
        "tint-sky": "hsl(var(--tint-sky))",
        "tint-steel": "hsl(var(--tint-steel))",
        "tint-periwinkle": "hsl(var(--tint-periwinkle))",

        // Semantic — mapped to Dell 1996 primitives
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      fontFamily: {
        display: ['"Arial Black"', 'Helvetica', 'system-ui', 'sans-serif'],
        heading: ['Helvetica', 'Arial', 'system-ui', 'sans-serif'],
        body: ['"Times New Roman"', 'Times', 'serif'],
        ui: ['Helvetica', 'Arial', 'system-ui', 'sans-serif'],
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
}
