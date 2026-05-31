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
    tint: "#b3bd95", // sage
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
    tint: "#8c9ae0", // periwinkle
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
    tint: "#a5b8c0", // steel
  },
];

export const Pricing = () => {
  return (
    <section id="pricing" className="bg-white">
      <div className="mx-auto max-w-[760px]">
        {/* Section eyebrow — peach */}
        <ScrollReveal width="100%">
          <div className="bg-[#e6915d] px-4 py-6">
            <p className="mb-1 font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
              Plans
            </p>
            <h2 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[36px] font-black uppercase leading-[1.0] text-black">
              CHOOSE YOUR<br />
              OPERATING PACE.
            </h2>
          </div>
        </ScrollReveal>

        {/* Pricing ribbon cards */}
        <div className="mt-0">
          {tiers.map((tier, index) => (
            <ScrollReveal key={index} width="100%" delay={index * 0.1}>
              <div className="border-x border-b border-black">
                {/* Card title bar */}
                <div className="border-b border-black bg-white px-3 py-1.5">
                  <div className="flex items-center justify-between">
                    <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                      {tier.name}
                    </h3>
                    <span className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-xl font-black text-black">
                      {tier.price}
                      {tier.price !== "Custom" && (
                        <span className="font-['Times_New_Roman',Times,serif] text-xs font-normal lowercase text-gray-600">
                          /mo
                        </span>
                      )}
                    </span>
                  </div>
                  {tier.popular && (
                    <div className="mt-1 inline-block border border-black bg-[#fcc20f] px-2 py-0">
                      <span className="font-[Helvetica,Arial,system-ui,sans-serif] text-[10px] font-bold uppercase tracking-[1px] text-black">
                        ★ RECOMMENDED ★
                      </span>
                    </div>
                  )}
                </div>
                {/* Card body — tinted */}
                <div className="px-4 py-3" style={{ backgroundColor: tier.tint }}>
                  <p className="mb-3 font-['Times_New_Roman',Times,serif] text-sm leading-[1.4] text-black">
                    {tier.description}
                  </p>
                  <ul className="space-y-1">
                    {tier.features.map((feature) => (
                      <li key={feature} className="flex items-center gap-2">
                        <Check className="h-3 w-3 shrink-0 text-black" />
                        <span className="font-['Times_New_Roman',Times,serif] text-sm text-black">
                          {feature}
                        </span>
                      </li>
                    ))}
                  </ul>
                  <div className="mt-3">
                    {tier.href.startsWith("mailto") ? (
                      <Button asChild>
                        <a href={tier.href}>{tier.cta}</a>
                      </Button>
                    ) : (
                      <Button asChild>
                        <Link to={tier.href}>{tier.cta}</Link>
                      </Button>
                    )}
                  </div>
                </div>
              </div>
            </ScrollReveal>
          ))}
        </div>

        {/* Bottom CTA — Dell red */}
        <ScrollReveal width="100%" delay={0.4}>
          <div className="border-x border-b border-black bg-[#e91d2a] p-4">
            <p className="mb-3 font-['Times_New_Roman',Times,serif] text-sm leading-[1.4] text-[#fffff0]">
              Put every application behind one gate. Keyles gives you the
              surface to configure, secure, and monitor every identity path.
            </p>
            <Button variant="outline" className="border-white text-white hover:bg-white hover:text-black" asChild>
              <Link to="/register">Start Deployment</Link>
            </Button>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
};
