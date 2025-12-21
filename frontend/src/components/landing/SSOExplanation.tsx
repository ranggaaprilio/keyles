import { ScrollReveal } from "./ScrollReveal";
import { Lock, ShieldCheck, LayoutGrid } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const steps = [
  {
    title: "One Login",
    description: "Users sign in once with a single set of credentials.",
    icon: Lock,
  },
  {
    title: "Secure Verification",
    description:
      "We verify identity with multi-factor authentication and security policies.",
    icon: ShieldCheck,
  },
  {
    title: "Access Everything",
    description:
      "Instant access to all authorized applications without re-entering passwords.",
    icon: LayoutGrid,
  },
];

export const SSOExplanation = () => {
  return (
    <section id="how-it-works" className="py-20 bg-background">
      <div className="container mx-auto px-4">
        <ScrollReveal width="100%">
          <h2 className="text-3xl font-bold text-center mb-12">How It Works</h2>
        </ScrollReveal>

        <div className="grid md:grid-cols-3 gap-8">
          {steps.map((step, index) => (
            <ScrollReveal key={index} width="100%" delay={index * 0.2}>
              <Card className="h-full text-center hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="mx-auto bg-primary/10 p-4 rounded-full mb-4 w-16 h-16 flex items-center justify-center">
                    <step.icon className="w-8 h-8 text-primary" />
                  </div>
                  <CardTitle>{step.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground">{step.description}</p>
                </CardContent>
              </Card>
            </ScrollReveal>
          ))}
        </div>
      </div>
    </section>
  );
};
