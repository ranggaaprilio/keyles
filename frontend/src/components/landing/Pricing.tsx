import { ArrowRight, Check } from "lucide-react";
import { Link } from "react-router-dom";
import { ScrollReveal } from "./ScrollReveal";

const tiers = [
  {
    name: "Starter",
    price: "$0",
    suffix: "/ month",
    description: "For small teams building their first secure identity layer.",
    features: ["Up to 50 users", "Core SSO", "1 tenant", "Email support"],
    cta: "Start free",
    href: "/register",
  },
  {
    name: "Pro",
    price: "$49",
    suffix: "/ month",
    description: "For growing organizations that need deeper control.",
    features: [
      "Up to 500 users",
      "Advanced policies",
      "Audit logs",
      "Priority support",
    ],
    cta: "Start free trial",
    href: "/register?plan=pro",
    popular: true,
  },
  {
    name: "Enterprise",
    price: "Custom",
    suffix: "",
    description: "For complex environments with tailored requirements.",
    features: [
      "Unlimited users",
      "Custom integrations",
      "Dedicated support",
      "SLA guarantees",
    ],
    cta: "Contact sales",
    href: "mailto:sales@keyles.com",
  },
];

const primaryCta =
  "inline-flex h-12 w-full items-center justify-center gap-2 rounded-full bg-primary text-sm font-medium text-primary-foreground no-underline transition-colors duration-fast hover:bg-primary/85 hover:text-primary-foreground";
const outlineCta =
  "inline-flex h-12 w-full items-center justify-center gap-2 rounded-full border border-input bg-background text-sm font-medium text-foreground no-underline transition-colors duration-fast hover:bg-muted hover:text-foreground";

export const Pricing = () => {
  return (
    <section id="pricing" className="bg-muted py-24 sm:py-32">
      <div className="mx-auto max-w-7xl px-5 sm:px-8 lg:px-12">
        <ScrollReveal width="100%">
          <div className="mx-auto max-w-2xl text-center">
            <p className="text-xs font-medium text-muted-foreground">
              Simple pricing
            </p>
            <h2 className="mt-5 text-4xl font-normal leading-[1.05] tracking-[-0.03em] text-foreground sm:text-5xl">
              Start small. Scale securely.
            </h2>
            <p className="mt-5 text-base leading-7 text-muted-foreground">
              Every plan includes the essentials for secure authentication.
              Move up when your team and access policies grow.
            </p>
          </div>
        </ScrollReveal>

        <div className="mt-14 grid gap-5 lg:grid-cols-3 lg:items-stretch">
          {tiers.map((tier, index) => (
            <ScrollReveal key={tier.name} width="100%" delay={index * 0.08}>
              <article
                className={`relative flex h-full flex-col rounded-xl border bg-background p-7 text-foreground shadow-card sm:p-8 ${
                  tier.popular ? "border-foreground ring-1 ring-foreground" : "border-border"
                }`}
              >
                {tier.popular && (
                  <span className="absolute right-6 top-6 rounded-full bg-primary px-3 py-1 text-xs font-medium text-primary-foreground">
                    Most popular
                  </span>
                )}
                <p className="text-sm font-medium text-muted-foreground">
                  {tier.name}
                </p>
                <div className="mt-8 flex items-end gap-2">
                  <span className="text-4xl font-normal tracking-[-0.03em]">
                    {tier.price}
                  </span>
                  {tier.suffix && (
                    <span className="pb-1 text-xs text-muted-foreground">
                      {tier.suffix}
                    </span>
                  )}
                </div>
                <p className="mt-5 min-h-[48px] text-sm leading-6 text-muted-foreground">
                  {tier.description}
                </p>

                <div className="my-7 h-px bg-border" />
                <ul className="space-y-4">
                  {tier.features.map((feature) => (
                    <li
                      key={feature}
                      className="flex items-center gap-3 text-sm text-foreground"
                    >
                      <span className="flex h-5 w-5 items-center justify-center rounded-full bg-success/10 text-success">
                        <Check className="h-3 w-3" />
                      </span>
                      {feature}
                    </li>
                  ))}
                </ul>

                <div className="mt-auto pt-9">
                  {tier.href.startsWith("mailto") ? (
                    <a href={tier.href} className={outlineCta}>
                      {tier.cta}
                      <ArrowRight className="h-4 w-4" />
                    </a>
                  ) : (
                    <Link
                      to={tier.href}
                      className={tier.popular ? primaryCta : outlineCta}
                    >
                      {tier.cta}
                      <ArrowRight className="h-4 w-4" />
                    </Link>
                  )}
                </div>
              </article>
            </ScrollReveal>
          ))}
        </div>

        <ScrollReveal width="100%" delay={0.15}>
          <div className="mt-20 overflow-hidden rounded-xl bg-primary px-7 py-12 text-primary-foreground sm:px-12 lg:flex lg:items-center lg:justify-between lg:px-16">
            <div>
              <p className="text-xs font-medium text-primary-foreground/60">
                Ready when you are
              </p>
              <h3 className="mt-4 max-w-2xl text-3xl font-normal leading-tight tracking-[-0.03em] sm:text-4xl">
                Give every application one trusted front door.
              </h3>
            </div>
            <Link
              to="/register"
              className="mt-8 inline-flex h-12 shrink-0 items-center justify-center gap-2 rounded-full bg-background px-7 text-sm font-medium text-foreground no-underline transition-colors duration-fast hover:bg-background/90 hover:text-foreground lg:mt-0"
            >
              Start your deployment
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
};
