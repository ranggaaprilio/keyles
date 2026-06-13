import {
  ArrowRight,
  Fingerprint,
  LayoutGrid,
  LockKeyhole,
  ShieldCheck,
  Users,
} from "lucide-react";
import { Link } from "react-router-dom";
import { ScrollReveal } from "./ScrollReveal";

const features = [
  {
    title: "One identity",
    description:
      "A single, secure sign-in gives people access to every approved application.",
    icon: Fingerprint,
    className: "bg-[#b3bd95]",
  },
  {
    title: "Policy by default",
    description:
      "Apply tenant, session, and OAuth policies from one predictable control plane.",
    icon: ShieldCheck,
    className: "bg-[#d77a7a]",
  },
  {
    title: "Roles that scale",
    description:
      "Map users to the right clients and permissions without access sprawl.",
    icon: Users,
    className: "bg-[#9ab6c8]",
  },
  {
    title: "Every app connected",
    description:
      "Standards-based OAuth 2.0 integration keeps implementation straightforward.",
    icon: LayoutGrid,
    className: "bg-[#8c9ae0]",
  },
];

export const SSOExplanation = () => {
  return (
    <section id="how-it-works" className="bg-[#f7f7f5] py-24 sm:py-32">
      <div className="mx-auto max-w-7xl px-5 sm:px-8 lg:px-12">
        <ScrollReveal width="100%">
          <div className="grid gap-8 lg:grid-cols-[0.8fr_1.2fr] lg:items-end">
            <div>
              <p className="font-ui text-xs font-bold uppercase tracking-[0.18em] text-[#e91d2a]">
                The Keyles platform
              </p>
              <h2 className="mt-5 max-w-lg font-heading text-4xl font-bold leading-[1.05] tracking-[-0.045em] text-black sm:text-5xl">
                Identity management without the noise.
              </h2>
            </div>
            <div className="lg:pb-1">
              <p className="max-w-xl font-ui text-base leading-7 text-black/55">
                Replace scattered login systems with one secure identity layer.
                Your team gets clear controls. Your users get a simple,
                consistent experience.
              </p>
              <Link
                to="/docs/oauth"
                className="mt-5 inline-flex items-center gap-2 font-ui text-sm font-bold text-black no-underline hover:text-[#e91d2a]"
              >
                See how integration works
                <ArrowRight className="h-4 w-4" />
              </Link>
            </div>
          </div>
        </ScrollReveal>

        <div className="mt-14 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {features.map((feature, index) => (
            <ScrollReveal key={feature.title} width="100%" delay={index * 0.08}>
              <article className="group flex min-h-[310px] flex-col rounded-2xl border border-black/10 bg-white p-6 transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_24px_60px_rgba(0,0,0,0.08)]">
                <div
                  className={`flex h-12 w-12 items-center justify-center rounded-xl ${feature.className}`}
                >
                  <feature.icon className="h-5 w-5 text-black" />
                </div>
                <div className="mt-auto pt-16">
                  <span className="font-ui text-xs font-bold text-black/30">
                    0{index + 1}
                  </span>
                  <h3 className="mt-3 font-heading text-xl font-bold tracking-[-0.025em]">
                    {feature.title}
                  </h3>
                  <p className="mt-3 font-ui text-sm leading-6 text-black/50">
                    {feature.description}
                  </p>
                </div>
              </article>
            </ScrollReveal>
          ))}
        </div>

        <ScrollReveal width="100%" delay={0.15}>
          <div className="mt-4 grid overflow-hidden rounded-2xl bg-black text-white lg:grid-cols-[1fr_0.9fr]">
            <div className="p-7 sm:p-10 lg:p-14">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#e91d2a]">
                <LockKeyhole className="h-5 w-5" />
              </div>
              <h3 className="mt-8 max-w-lg font-heading text-3xl font-bold leading-tight tracking-[-0.04em] sm:text-4xl">
                Secure access becomes the default, not another project.
              </h3>
              <p className="mt-5 max-w-xl font-ui text-sm leading-6 text-white/55">
                Centralize authentication once, then let every client inherit
                the same reliable security model.
              </p>
            </div>
            <div className="grid border-t border-white/10 sm:grid-cols-3 lg:grid-cols-1 lg:border-l lg:border-t-0">
              {[
                ["99.98%", "Authentication success"],
                ["<100ms", "Policy decisions"],
                ["24/7", "Session visibility"],
              ].map(([value, label]) => (
                <div
                  key={label}
                  className="border-b border-white/10 p-7 last:border-b-0 sm:border-b-0 sm:border-r sm:last:border-r-0 lg:border-b lg:border-r-0"
                >
                  <p className="font-heading text-3xl font-bold text-[#d77a7a]">
                    {value}
                  </p>
                  <p className="mt-2 font-ui text-xs text-white/45">{label}</p>
                </div>
              ))}
            </div>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
};
