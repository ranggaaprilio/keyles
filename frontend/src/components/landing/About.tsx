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
    <section id="about" className="overflow-hidden bg-background py-24 sm:py-32">
      <div className="mx-auto grid max-w-7xl gap-16 px-5 sm:px-8 lg:grid-cols-2 lg:items-center lg:px-12">
        <ScrollReveal width="100%">
          <div className="relative mx-auto w-full max-w-xl">
            <div className="relative rounded-xl border border-border bg-card p-5 shadow-elevated sm:p-8">
              <div className="flex items-center justify-between border-b border-border pb-6">
                <div>
                  <p className="text-xs text-muted-foreground">Security center</p>
                  <h3 className="mt-1 text-xl font-medium tracking-tight">
                    Tenant protection
                  </h3>
                </div>
                <span className="flex items-center gap-2 rounded-full bg-success/10 px-3 py-1.5 text-xs font-medium text-success">
                  <span className="h-2 w-2 rounded-full bg-success" />
                  Healthy
                </span>
              </div>

              <div className="my-7 rounded-lg bg-muted p-5">
                <div className="flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary text-primary-foreground">
                    <Shield className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">
                      Security posture
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      All recommended controls enabled
                    </p>
                  </div>
                </div>
                <div className="mt-6 h-2 overflow-hidden rounded-full bg-border">
                  <div className="h-full w-[94%] rounded-full bg-success" />
                </div>
                <div className="mt-3 flex justify-between text-[11px] text-muted-foreground">
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
                    className="flex items-center justify-between rounded-md border border-border px-4 py-3"
                  >
                    <span className="flex items-center gap-3 text-xs text-foreground">
                      <CheckCircle2 className="h-4 w-4 text-success" />
                      {label}
                    </span>
                    <span className="text-[11px] text-muted-foreground">
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
            <p className="text-xs font-medium text-muted-foreground">
              Security that stays visible
            </p>
            <h2 className="mt-5 max-w-xl text-4xl font-normal leading-[1.05] tracking-[-0.03em] text-foreground sm:text-5xl">
              Strict by design. Clear in practice.
            </h2>
            <p className="mt-6 max-w-xl text-base leading-7 text-muted-foreground">
              Security should be easy to inspect and difficult to bypass.
              Keyles keeps tenant controls, active sessions, client access, and
              audit activity visible without burying your team in complexity.
            </p>

            <div className="mt-10 space-y-7">
              {controls.map((control) => (
                <div key={control.title} className="flex gap-4">
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-muted text-foreground">
                    <control.icon className="h-5 w-5" />
                  </div>
                  <div>
                    <h3 className="text-base font-medium text-foreground">
                      {control.title}
                    </h3>
                    <p className="mt-1 text-sm leading-6 text-muted-foreground">
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
