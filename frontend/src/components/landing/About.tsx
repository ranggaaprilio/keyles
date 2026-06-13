import { Activity, CheckCircle2, KeyRound, Shield, UserCheck } from "lucide-react";
import { ScrollReveal } from "./ScrollReveal";

const controls = [
  {
    title: "Session visibility",
    description: "Review and revoke active sessions across every tenant.",
    icon: Activity,
  },
  {
    title: "Role-based access",
    description: "Keep permissions explicit as teams and applications grow.",
    icon: UserCheck,
  },
  {
    title: "OAuth client control",
    description: "Manage credentials, redirects, and grants from one place.",
    icon: KeyRound,
  },
];

export const About = () => {
  return (
    <section id="about" className="overflow-hidden bg-white py-24 sm:py-32">
      <div className="mx-auto grid max-w-7xl gap-16 px-5 sm:px-8 lg:grid-cols-2 lg:items-center lg:px-12">
        <ScrollReveal width="100%">
          <div className="relative mx-auto w-full max-w-xl">
            <div className="absolute -left-10 -top-10 h-40 w-40 rounded-full bg-[#d77a7a]/25 blur-3xl" />
            <div className="relative rounded-[1.75rem] bg-black p-5 text-white shadow-[0_32px_80px_rgba(0,0,0,0.18)] sm:p-8">
              <div className="flex items-center justify-between border-b border-white/10 pb-6">
                <div>
                  <p className="font-ui text-xs text-white/40">Security center</p>
                  <h3 className="mt-1 font-heading text-xl font-bold">
                    Tenant protection
                  </h3>
                </div>
                <span className="flex items-center gap-2 rounded-full bg-[#b3bd95]/15 px-3 py-2 font-ui text-xs font-semibold text-[#b3bd95]">
                  <span className="h-2 w-2 rounded-full bg-[#b3bd95]" />
                  Healthy
                </span>
              </div>

              <div className="my-7 rounded-2xl bg-white/[0.05] p-5">
                <div className="flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#d77a7a] text-black">
                    <Shield className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="font-ui text-sm font-semibold">
                      Security posture
                    </p>
                    <p className="mt-1 font-ui text-xs text-white/40">
                      All recommended controls enabled
                    </p>
                  </div>
                </div>
                <div className="mt-6 h-2 overflow-hidden rounded-full bg-white/10">
                  <div className="h-full w-[94%] rounded-full bg-[#b3bd95]" />
                </div>
                <div className="mt-3 flex justify-between font-ui text-[11px] text-white/35">
                  <span>Protection score</span>
                  <span>94 / 100</span>
                </div>
              </div>

              <div className="space-y-3">
                {[
                  ["TLS and secure headers", "Enabled"],
                  ["Rate limit protection", "Active"],
                  ["Audit event logging", "Streaming"],
                ].map(([label, status]) => (
                  <div
                    key={label}
                    className="flex items-center justify-between rounded-xl border border-white/10 px-4 py-4"
                  >
                    <span className="flex items-center gap-3 font-ui text-xs text-white/70">
                      <CheckCircle2 className="h-4 w-4 text-[#b3bd95]" />
                      {label}
                    </span>
                    <span className="font-ui text-[11px] text-white/35">
                      {status}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </ScrollReveal>

        <ScrollReveal width="100%" delay={0.1}>
          <div>
            <p className="font-ui text-xs font-bold uppercase tracking-[0.18em] text-[#e91d2a]">
              Security that stays visible
            </p>
            <h2 className="mt-5 max-w-xl font-heading text-4xl font-bold leading-[1.05] tracking-[-0.045em] sm:text-5xl">
              Strict by design. Clear in practice.
            </h2>
            <p className="mt-6 max-w-xl font-ui text-base leading-7 text-black/55">
              Security should be easy to inspect and difficult to bypass.
              Keyles keeps tenant controls, active sessions, client access, and
              audit activity visible without burying your team in complexity.
            </p>

            <div className="mt-10 space-y-7">
              {controls.map((control) => (
                <div key={control.title} className="flex gap-4">
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#f3f3f0]">
                    <control.icon className="h-5 w-5 text-black" />
                  </div>
                  <div>
                    <h3 className="font-heading text-base font-bold">
                      {control.title}
                    </h3>
                    <p className="mt-1 font-ui text-sm leading-6 text-black/50">
                      {control.description}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
};
