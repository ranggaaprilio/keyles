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
  },
  {
    title: "Policy by default",
    description:
      "Apply tenant, session, and OAuth policies from one predictable control plane.",
    icon: ShieldCheck,
  },
  {
    title: "Roles that scale",
    description:
      "Map users to the right clients and permissions without access sprawl.",
    icon: Users,
  },
  {
    title: "Every app connected",
    description:
      "Standards-based OAuth 2.0 integration keeps implementation straightforward.",
    icon: LayoutGrid,
  },
];

export const SSOExplanation = () => {
  return (
    <section id="how-it-works" className="bg-muted py-24 sm:py-32">
      <div className="mx-auto max-w-7xl px-5 sm:px-8 lg:px-12">
        <ScrollReveal width="100%">
          <div className="grid gap-8 lg:grid-cols-[0.8fr_1.2fr] lg:items-end">
            <div>
              <p className="text-xs font-medium text-muted-foreground">
                The Keyles platform
              </p>
              <h2 className="mt-5 max-w-lg text-4xl font-normal leading-[1.05] tracking-[-0.03em] text-foreground sm:text-5xl">
                Identity management without the noise.
              </h2>
            </div>
            <div className="lg:pb-1">
              <p className="max-w-xl text-base leading-7 text-muted-foreground">
                Replace scattered login systems with one secure identity layer.
                Your team gets clear controls. Your users get a simple,
                consistent experience.
              </p>
              <Link
                to="/docs/oauth"
                className="mt-5 inline-flex items-center gap-2 text-sm font-medium text-foreground no-underline transition-colors duration-fast hover:text-muted-foreground"
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
              <article className="group flex min-h-[310px] flex-col rounded-xl border border-border bg-background p-6 shadow-card transition-all duration-base hover:-translate-y-1 hover:shadow-elevated">
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-foreground">
                  <feature.icon className="h-5 w-5" />
                </div>
                <div className="mt-auto pt-16">
                  <span className="text-xs font-medium text-muted-foreground">
                    0{index + 1}
                  </span>
                  <h3 className="mt-3 text-xl font-medium tracking-tight text-foreground">
                    {feature.title}
                  </h3>
                  <p className="mt-3 text-sm leading-6 text-muted-foreground">
                    {feature.description}
                  </p>
                </div>
              </article>
            </ScrollReveal>
          ))}
        </div>

        <ScrollReveal width="100%" delay={0.15}>
          <div className="mt-4 grid overflow-hidden rounded-xl bg-primary text-primary-foreground lg:grid-cols-[1fr_0.9fr]">
            <div className="p-7 sm:p-10 lg:p-14">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary-foreground/10">
                <LockKeyhole className="h-5 w-5" />
              </div>
              <h3 className="mt-8 max-w-lg text-3xl font-normal leading-tight tracking-[-0.03em] sm:text-4xl">
                Secure access becomes the default, not another project.
              </h3>
              <p className="mt-5 max-w-xl text-sm leading-6 text-primary-foreground/60">
                Centralize authentication once, then let every client inherit
                the same reliable security model.
              </p>
            </div>
            <div className="grid border-t border-primary-foreground/10 sm:grid-cols-3 lg:grid-cols-1 lg:border-l lg:border-t-0">
              {[
                ["99.98%", "Authentication success"],
                ["<100ms", "Policy decisions"],
                ["24/7", "Session visibility"],
              ].map(([value, label]) => (
                <div
                  key={label}
                  className="border-b border-primary-foreground/10 p-7 last:border-b-0 sm:border-b-0 sm:border-r sm:last:border-r-0 lg:border-b lg:border-r-0"
                >
                  <p className="text-3xl font-normal tracking-tight">
                    {value}
                  </p>
                  <p className="mt-2 text-xs text-primary-foreground/60">{label}</p>
                </div>
              ))}
            </div>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
};
