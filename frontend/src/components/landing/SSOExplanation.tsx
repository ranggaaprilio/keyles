import { ScrollReveal } from "./ScrollReveal";
import { Lock, ShieldCheck, LayoutGrid } from "lucide-react";

const steps = [
  {
    title: "One identity",
    description:
      "Users authenticate once, then move across approved applications without password repetition.",
    icon: Lock,
    tint: "#b3bd95", // sage
  },
  {
    title: "Policy enforced",
    description:
      "Tenant settings, OAuth clients, and session controls stay visible to administrators.",
    icon: ShieldCheck,
    tint: "#d77a7a", // salmon
  },
  {
    title: "Access mapped",
    description:
      "Roles and client grants keep each application tied to the right user boundary.",
    icon: LayoutGrid,
    tint: "#9ab6c8", // sky
  },
];

export const SSOExplanation = () => {
  return (
    <section id="how-it-works" className="bg-white">
      <div className="mx-auto max-w-[760px]">
        {/* Section eyebrow — periwinkle */}
        <ScrollReveal width="100%">
          <div className="bg-[#8c9ae0] px-4 py-6">
            <p className="mb-1 font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
              Platform Sequence
            </p>
            <h2 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[36px] font-black uppercase leading-[1.0] text-black">
              HOW IT WORKS
            </h2>
          </div>
        </ScrollReveal>

        {/* Ribbon cards */}
        <div className="mt-0">
          {steps.map((step, index) => (
            <ScrollReveal key={index} width="100%" delay={index * 0.15}>
              <div className="border-x border-b border-black">
                {/* Card title bar */}
                <div className="flex items-center gap-2 border-b border-black bg-white px-3 py-1.5">
                  <step.icon className="h-4 w-4 text-black" />
                  <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                    {step.title}
                  </h3>
                </div>
                {/* Card body — tinted */}
                <div className="px-4 py-3" style={{ backgroundColor: step.tint }}>
                  <p className="font-['Times_New_Roman',Times,serif] text-sm leading-[1.4] text-black">
                    {step.description}
                  </p>
                </div>
              </div>
            </ScrollReveal>
          ))}
        </div>
      </div>
    </section>
  );
};
