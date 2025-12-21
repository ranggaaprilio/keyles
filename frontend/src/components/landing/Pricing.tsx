import { ScrollReveal } from "./ScrollReveal";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
    <section id="pricing" className="py-20 bg-background">
      <div className="container mx-auto px-4">
        <ScrollReveal width="100%">
          <h2 className="text-3xl font-bold text-center mb-4">
            Simple, Transparent Pricing
          </h2>
          <p className="text-center text-muted-foreground mb-12 max-w-2xl mx-auto">
            Choose the plan that fits your needs. No hidden fees.
          </p>
        </ScrollReveal>

        <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          {tiers.map((tier, index) => (
            <ScrollReveal key={index} width="100%" delay={index * 0.1 + 0.2}>
              <Card
                className={`h-full flex flex-col ${tier.popular ? "border-primary shadow-lg relative" : ""}`}
              >
                {tier.popular && (
                  <div className="absolute top-0 right-0 bg-primary text-primary-foreground text-xs font-bold px-3 py-1 rounded-bl-lg rounded-tr-lg">
                    POPULAR
                  </div>
                )}
                <CardHeader>
                  <CardTitle className="text-2xl">{tier.name}</CardTitle>
                  <CardDescription>{tier.description}</CardDescription>
                </CardHeader>
                <CardContent className="flex-grow">
                  <div className="text-4xl font-bold mb-6">
                    {tier.price}
                    {tier.price !== "Custom" && (
                      <span className="text-lg font-normal text-muted-foreground">
                        /mo
                      </span>
                    )}
                  </div>
                  <ul className="space-y-3">
                    {tier.features.map((feature, i) => (
                      <li key={i} className="flex items-center gap-2">
                        <Check className="w-4 h-4 text-primary" />
                        <span className="text-sm text-muted-foreground">
                          {feature}
                        </span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
                <CardFooter>
                  <Button className="w-full" variant={tier.variant} asChild>
                    {tier.href.startsWith("mailto") ? (
                      <a href={tier.href}>{tier.cta}</a>
                    ) : (
                      <Link to={tier.href}>{tier.cta}</Link>
                    )}
                  </Button>
                </CardFooter>
              </Card>
            </ScrollReveal>
          ))}
        </div>
      </div>
    </section>
  );
};
