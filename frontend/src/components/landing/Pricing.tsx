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

export const Pricing = () => {
  return (
    <section id="pricing" className="bg-[#f7f7f5] py-24 sm:py-32">
      <div className="mx-auto max-w-7xl px-5 sm:px-8 lg:px-12">
        <ScrollReveal width="100%">
          <div className="mx-auto max-w-2xl text-center">
            <p className="font-ui text-xs font-bold uppercase tracking-[0.18em] text-[#e91d2a]">
              Simple pricing
            </p>
            <h2 className="mt-5 font-heading text-4xl font-bold leading-[1.05] tracking-[-0.045em] sm:text-5xl">
              Start small. Scale securely.
            </h2>
            <p className="mt-5 font-ui text-base leading-7 text-black/50">
              Every plan includes the essentials for secure authentication.
              Move up when your team and access policies grow.
            </p>
          </div>
        </ScrollReveal>

        <div className="mt-14 grid gap-5 lg:grid-cols-3 lg:items-stretch">
          {tiers.map((tier, index) => (
            <ScrollReveal key={tier.name} width="100%" delay={index * 0.08}>
              <article
                className={`relative flex h-full flex-col rounded-2xl border p-7 sm:p-8 ${
                  tier.popular
                    ? "border-black bg-black text-white shadow-[0_28px_70px_rgba(0,0,0,0.16)]"
                    : "border-black/10 bg-white text-black"
                }`}
              >
                {tier.popular && (
                  <span className="absolute right-6 top-6 rounded-full bg-[#d77a7a] px-3 py-1.5 font-ui text-[10px] font-bold uppercase tracking-[0.12em] text-black">
                    Most popular
                  </span>
                )}
                <p
                  className={`font-ui text-sm font-bold ${
                    tier.popular ? "text-white/55" : "text-black/45"
                  }`}
                >
                  {tier.name}
                </p>
                <div className="mt-8 flex items-end gap-2">
                  <span className="font-heading text-4xl font-bold tracking-[-0.04em]">
                    {tier.price}
                  </span>
                  {tier.suffix && (
                    <span
                      className={`pb-1 font-ui text-xs ${
                        tier.popular ? "text-white/35" : "text-black/35"
                      }`}
                    >
                      {tier.suffix}
                    </span>
                  )}
                </div>
                <p
                  className={`mt-5 min-h-[48px] font-ui text-sm leading-6 ${
                    tier.popular ? "text-white/50" : "text-black/50"
                  }`}
                >
                  {tier.description}
                </p>

                <div
                  className={`my-7 h-px ${
                    tier.popular ? "bg-white/10" : "bg-black/10"
                  }`}
                />
                <ul className="space-y-4">
                  {tier.features.map((feature) => (
                    <li
                      key={feature}
                      className={`flex items-center gap-3 font-ui text-sm ${
                        tier.popular ? "text-white/70" : "text-black/65"
                      }`}
                    >
                      <span
                        className={`flex h-5 w-5 items-center justify-center rounded-full ${
                          tier.popular ? "bg-[#b3bd95]" : "bg-[#b3bd95]/55"
                        }`}
                      >
                        <Check className="h-3 w-3 text-black" />
                      </span>
                      {feature}
                    </li>
                  ))}
                </ul>

                <div className="mt-auto pt-9">
                  {tier.href.startsWith("mailto") ? (
                    <a
                      href={tier.href}
                      className="inline-flex h-12 w-full items-center justify-center gap-2 rounded-xl border border-black/15 font-ui text-sm font-bold text-black no-underline transition-colors hover:bg-black hover:text-white"
                    >
                      {tier.cta}
                      <ArrowRight className="h-4 w-4" />
                    </a>
                  ) : (
                    <Link
                      to={tier.href}
                      className={`inline-flex h-12 w-full items-center justify-center gap-2 rounded-xl font-ui text-sm font-bold no-underline transition-all hover:-translate-y-0.5 ${
                        tier.popular
                          ? "bg-[#e91d2a] text-white hover:text-white"
                          : "border border-black/15 text-black hover:bg-black hover:text-white"
                      }`}
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
          <div className="mt-20 overflow-hidden rounded-[1.75rem] bg-[#e91d2a] px-7 py-12 text-white sm:px-12 lg:flex lg:items-center lg:justify-between lg:px-16">
            <div>
              <p className="font-ui text-xs font-bold uppercase tracking-[0.18em] text-white/65">
                Ready when you are
              </p>
              <h3 className="mt-4 max-w-2xl font-heading text-3xl font-bold leading-tight tracking-[-0.04em] sm:text-4xl">
                Give every application one trusted front door.
              </h3>
            </div>
            <Link
              to="/register"
              className="mt-8 inline-flex h-13 shrink-0 items-center justify-center gap-2 rounded-xl bg-white px-6 py-4 font-ui text-sm font-bold text-black no-underline transition-transform hover:-translate-y-0.5 hover:text-black lg:mt-0"
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
