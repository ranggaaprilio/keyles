import { ScrollReveal } from "./ScrollReveal";
import { Lock, ShieldCheck, LayoutGrid } from "lucide-react";

const steps = [
  {
    title: "One identity",
    description:
      "Users authenticate once, then move across approved applications without password repetition.",
    icon: Lock,
  },
  {
    title: "Policy enforced",
    description:
      "Tenant settings, OAuth clients, and session controls stay visible to administrators.",
    icon: ShieldCheck,
  },
  {
    title: "Access mapped",
    description:
      "Roles and client grants keep each application tied to the right user boundary.",
    icon: LayoutGrid,
  },
];

export const SSOExplanation = () => {
  return (
    <section id="how-it-works" className="border-b border-[#3c3c3c] bg-black py-24">
      <div className="mx-auto max-w-[1440px] px-5 md:px-8">
        <ScrollReveal width="100%">
          <div className="mb-12 max-w-3xl">
            <p className="mb-3 text-sm font-bold uppercase tracking-[1.5px] text-[#bbbbbb]">
              Platform sequence
            </p>
            <h2 className="text-4xl font-bold uppercase leading-tight text-white md:text-6xl">
              Authentication without drift.
            </h2>
          </div>
        </ScrollReveal>

        <div className="grid gap-6 md:grid-cols-3">
          {steps.map((step, index) => (
            <ScrollReveal key={index} width="100%" delay={index * 0.2}>
              <article className="h-full border border-[#3c3c3c] bg-[#0d0d0d] p-6">
                <div className="mb-8 flex h-12 w-12 items-center justify-center rounded-full bg-[#1a1a1a] text-white">
                  <step.icon className="h-5 w-5" />
                </div>
                <p className="mb-3 text-xs font-bold uppercase tracking-[1.5px] text-[#7e7e7e]">
                  0{index + 1}
                </p>
                <h3 className="mb-4 text-2xl font-bold uppercase leading-tight text-white">
                  {step.title}
                </h3>
                <p className="text-base font-light leading-7 text-[#bbbbbb]">
                  {step.description}
                </p>
              </article>
            </ScrollReveal>
          ))}
        </div>
      </div>
    </section>
  );
};
