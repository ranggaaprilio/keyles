import { ScrollReveal } from "./ScrollReveal";
import { Button } from "@/components/ui/button";
import { Check } from "lucide-react";
import { Link } from "react-router-dom";

const tiers = [
  {
    name: "Starter",
    price: "$0",
    description: "Perfect for small teams getting started.",
    features: ["Up to 50 users", "Basic SSO", "Email Support", "1 Tenant"],
    cta: "Get Started",
    href: "/register",
    variant: "outline" as const,
  },
  {
    name: "Pro",
    price: "$49",
    description: "For growing businesses with advanced needs.",
    features: [
      "Up to 500 users",
      "Advanced SSO Policies",
      "Priority Support",
      "Audit Logs",
    ],
    cta: "Start Free Trial",
    href: "/register?plan=pro",
    variant: "default" as const,
    popular: true,
  },
  {
    name: "Enterprise",
    price: "Custom",
    description: "Tailored solutions for large organizations.",
    features: [
      "Unlimited users",
      "Dedicated Success Manager",
      "SLA Guarantees",
      "Custom Integrations",
    ],
    cta: "Contact Sales",
    href: "mailto:sales@keyles.com",
    variant: "outline" as const,
  },
];

export const Pricing = () => {
  return (
    <section id="pricing" className="border-y border-[#3c3c3c] bg-black py-24">
      <div className="mx-auto max-w-[1440px] px-5 md:px-8">
        <ScrollReveal width="100%">
          <div className="mb-12 grid gap-6 md:grid-cols-[1fr_420px] md:items-end">
            <div>
              <p className="mb-3 text-sm font-bold uppercase tracking-[1.5px] text-[#bbbbbb]">
                Plans
              </p>
              <h2 className="text-4xl font-bold uppercase leading-tight text-white md:text-6xl">
                Choose the operating pace.
              </h2>
            </div>
            <p className="text-base font-light leading-7 text-[#bbbbbb]">
              Start small, then expand users, clients, audit visibility, and
              support as your identity surface grows.
            </p>
          </div>
        </ScrollReveal>

        <div className="grid gap-6 md:grid-cols-3">
          {tiers.map((tier, index) => (
            <ScrollReveal key={index} width="100%" delay={index * 0.1 + 0.2}>
              <article className="relative flex h-full flex-col border border-[#3c3c3c] bg-[#0d0d0d] p-6">
                {tier.popular && (
                  <div className="absolute left-0 right-0 top-0 grid h-1 grid-cols-3">
                    <span className="bg-[#0066b1]" />
                    <span className="bg-[#1c69d4]" />
                    <span className="bg-[#e22718]" />
                  </div>
                )}
                <div className="pb-8">
                  <p className="mb-3 text-sm font-bold uppercase tracking-[1.5px] text-[#7e7e7e]">
                    {tier.popular ? "Recommended" : "Tier"}
                  </p>
                  <h3 className="text-3xl font-bold uppercase text-white">
                    {tier.name}
                  </h3>
                  <p className="mt-4 min-h-14 text-sm font-light leading-6 text-[#bbbbbb]">
                    {tier.description}
                  </p>
                </div>

                <div className="border-y border-[#3c3c3c] py-8">
                  <div className="text-5xl font-bold uppercase leading-none text-white">
                    {tier.price}
                    {tier.price !== "Custom" && (
                      <span className="ml-2 text-base font-light lowercase text-[#bbbbbb]">
                        /mo
                      </span>
                    )}
                  </div>
                </div>

                <ul className="flex-grow space-y-4 py-8">
                  {tier.features.map((feature) => (
                    <li key={feature} className="flex items-start gap-3">
                      <Check className="mt-1 h-4 w-4 shrink-0 text-white" />
                      <span className="text-sm font-light leading-6 text-[#bbbbbb]">
                        {feature}
                      </span>
                    </li>
                  ))}
                </ul>

                <Button className="w-full" variant={tier.variant} asChild>
                  {tier.href.startsWith("mailto") ? (
                    <a href={tier.href}>{tier.cta}</a>
                  ) : (
                    <Link to={tier.href}>{tier.cta}</Link>
                  )}
                </Button>
              </article>
            </ScrollReveal>
          ))}
        </div>
      </div>

      <div
        className="mt-24 min-h-[420px] bg-cover bg-center"
        style={{
          backgroundImage:
            "linear-gradient(180deg, rgba(0,0,0,.15), rgba(0,0,0,.7)), url('https://images.unsplash.com/photo-1541443131876-44b03de101c5?auto=format&fit=crop&w=2400&q=85')",
        }}
      >
        <div className="mx-auto flex min-h-[420px] max-w-[1440px] flex-col items-center justify-center px-5 py-20 text-center md:px-8">
          <ScrollReveal width="100%">
            <h2 className="mx-auto max-w-3xl text-4xl font-bold uppercase leading-tight text-white md:text-6xl">
              Put every application behind one gate.
            </h2>
          </ScrollReveal>
          <ScrollReveal width="100%" delay={0.2}>
            <div className="mt-8">
              <Button size="lg" variant="outline" asChild>
                <Link to="/register">Start deployment</Link>
              </Button>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>
  );
};
