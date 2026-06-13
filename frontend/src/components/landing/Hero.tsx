import { Link } from "react-router-dom";
import {
  ArrowRight,
  Check,
  ChevronDown,
  Fingerprint,
  MoreHorizontal,
  ShieldCheck,
  Users,
} from "lucide-react";
import { isAuthenticated } from "@/services/api/auth";
import { ScrollReveal } from "./ScrollReveal";

const integrations = ["ACME", "LUMEN", "NORTH", "ORBIT", "APEX"];

export const Hero = () => {
  const isLoggedIn = isAuthenticated();

  return (
    <section className="relative overflow-hidden bg-black text-white">
      <div className="absolute inset-0 landing-grid opacity-40" />
      <div className="absolute -right-40 top-12 h-[520px] w-[520px] rounded-full bg-[#e91d2a]/15 blur-[120px]" />
      <div className="absolute -left-40 bottom-0 h-[420px] w-[420px] rounded-full bg-[#8c9ae0]/15 blur-[120px]" />

      <div className="relative mx-auto grid max-w-7xl items-center gap-16 px-5 pb-20 pt-16 sm:px-8 sm:pb-28 sm:pt-24 lg:grid-cols-[0.9fr_1.1fr] lg:px-12 lg:py-32">
        <ScrollReveal width="100%">
          <div>
            <div className="mb-7 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/[0.06] px-4 py-2 font-ui text-xs font-semibold text-white/80">
              <span className="h-2 w-2 rounded-full bg-[#b3bd95]" />
              Identity infrastructure, simplified
            </div>
            <h1 className="max-w-2xl font-heading text-5xl font-bold leading-[0.98] tracking-[-0.055em] text-white sm:text-6xl lg:text-7xl">
              One login.
              <br />
              Every app.
              <br />
              <span className="text-[#d77a7a]">Total control.</span>
            </h1>
            <p className="mt-7 max-w-xl font-ui text-base leading-7 text-white/60 sm:text-lg">
              Give every team secure access to the tools they need. Keyles
              brings SSO, roles, sessions, and OAuth clients into one clear
              operating surface.
            </p>
            <div className="mt-9 flex flex-col gap-3 sm:flex-row">
              <Link
                to={isLoggedIn ? "/dashboard" : "/register"}
                className="inline-flex h-13 items-center justify-center gap-2 rounded-xl bg-[#e91d2a] px-6 py-4 font-ui text-sm font-bold text-white no-underline shadow-[0_16px_40px_rgba(233,29,42,0.25)] transition-all hover:-translate-y-0.5 hover:text-white hover:shadow-[0_20px_50px_rgba(233,29,42,0.35)]"
              >
                {isLoggedIn ? "Open dashboard" : "Start for free"}
                <ArrowRight className="h-4 w-4" />
              </Link>
              <Link
                to="/docs/oauth"
                className="inline-flex h-13 items-center justify-center rounded-xl border border-white/20 bg-white/[0.04] px-6 py-4 font-ui text-sm font-bold text-white no-underline transition-colors hover:bg-white/10 hover:text-white"
              >
                Explore documentation
              </Link>
            </div>
            <div className="mt-8 flex flex-wrap gap-x-6 gap-y-3 font-ui text-xs text-white/50">
              {["No credit card", "Deploy in minutes", "OAuth 2.0 ready"].map(
                (item) => (
                  <span key={item} className="flex items-center gap-2">
                    <Check className="h-3.5 w-3.5 text-[#b3bd95]" />
                    {item}
                  </span>
                ),
              )}
            </div>
          </div>
        </ScrollReveal>

        <ScrollReveal width="100%" delay={0.15}>
          <div className="relative mx-auto w-full max-w-2xl lg:mx-0">
            <div className="absolute -inset-6 rounded-[2rem] bg-gradient-to-br from-[#8c9ae0]/20 via-transparent to-[#e91d2a]/20 blur-2xl" />
            <div className="relative overflow-hidden rounded-[1.5rem] border border-white/15 bg-[#111111] shadow-2xl">
              <div className="flex items-center justify-between border-b border-white/10 px-5 py-4">
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full bg-[#e91d2a]" />
                  <span className="h-2.5 w-2.5 rounded-full bg-[#fcc20f]" />
                  <span className="h-2.5 w-2.5 rounded-full bg-[#b3bd95]" />
                </div>
                <span className="font-ui text-xs text-white/35">
                  console.keyles.io
                </span>
                <MoreHorizontal className="h-4 w-4 text-white/35" />
              </div>

              <div className="grid min-h-[430px] grid-cols-[72px_1fr] sm:grid-cols-[176px_1fr]">
                <aside className="border-r border-white/10 bg-black/30 p-3 sm:p-4">
                  <div className="mb-7 flex items-center gap-2 px-2">
                    <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-[#e91d2a]">
                      <Fingerprint className="h-4 w-4 text-white" />
                    </span>
                    <span className="hidden font-heading text-sm font-bold sm:block">
                      Keyles
                    </span>
                  </div>
                  <div className="space-y-2">
                    {[
                      { label: "Overview", icon: ShieldCheck, active: true },
                      { label: "Users", icon: Users },
                      { label: "Clients", icon: Fingerprint },
                    ].map((item) => (
                      <div
                        key={item.label}
                        className={`flex items-center gap-3 rounded-lg px-2.5 py-2.5 ${
                          item.active
                            ? "bg-white/10 text-white"
                            : "text-white/35"
                        }`}
                      >
                        <item.icon className="h-4 w-4 shrink-0" />
                        <span className="hidden font-ui text-xs sm:block">
                          {item.label}
                        </span>
                      </div>
                    ))}
                  </div>
                </aside>

                <div className="min-w-0 p-4 sm:p-6">
                  <div className="mb-6 flex items-center justify-between">
                    <div>
                      <p className="font-ui text-xs text-white/40">Overview</p>
                      <h2 className="mt-1 font-heading text-lg font-bold sm:text-xl">
                        Good morning, Alex
                      </h2>
                    </div>
                    <button
                      type="button"
                      className="flex items-center gap-2 rounded-lg border border-white/10 px-3 py-2 font-ui text-xs text-white/60"
                    >
                      ACME Inc.
                      <ChevronDown className="h-3 w-3" />
                    </button>
                  </div>

                  <div className="grid gap-3 sm:grid-cols-2">
                    <div className="rounded-xl border border-white/10 bg-white/[0.04] p-4">
                      <div className="flex items-start justify-between">
                        <span className="font-ui text-xs text-white/45">
                          Active users
                        </span>
                        <Users className="h-4 w-4 text-[#8c9ae0]" />
                      </div>
                      <p className="mt-5 font-heading text-3xl font-bold">2,418</p>
                      <p className="mt-1 font-ui text-[11px] text-[#b3bd95]">
                        +12.4% this month
                      </p>
                    </div>
                    <div className="rounded-xl border border-white/10 bg-white/[0.04] p-4">
                      <div className="flex items-start justify-between">
                        <span className="font-ui text-xs text-white/45">
                          Success rate
                        </span>
                        <ShieldCheck className="h-4 w-4 text-[#d77a7a]" />
                      </div>
                      <p className="mt-5 font-heading text-3xl font-bold">
                        99.98%
                      </p>
                      <p className="mt-1 font-ui text-[11px] text-white/35">
                        Last 30 days
                      </p>
                    </div>
                  </div>

                  <div className="mt-3 rounded-xl border border-white/10 bg-white/[0.04] p-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-ui text-xs font-semibold">
                          Authentication activity
                        </p>
                        <p className="mt-1 font-ui text-[11px] text-white/35">
                          Successful sign-ins
                        </p>
                      </div>
                      <span className="font-ui text-xs text-white/40">7 days</span>
                    </div>
                    <div className="mt-6 flex h-24 items-end gap-2">
                      {[40, 65, 48, 78, 60, 90, 72, 98, 84, 105, 94, 118].map(
                        (height, index) => (
                          <div
                            key={`${height}-${index}`}
                            className="flex-1 rounded-t-sm bg-[#e91d2a]/80"
                            style={{ height: `${Math.min(height, 100)}%` }}
                          />
                        ),
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </ScrollReveal>
      </div>

      <div className="relative border-t border-white/10">
        <div className="mx-auto flex max-w-7xl flex-col gap-6 px-5 py-8 sm:px-8 lg:flex-row lg:items-center lg:px-12">
          <p className="shrink-0 font-ui text-xs font-semibold uppercase tracking-[0.18em] text-white/35">
            Built for modern teams
          </p>
          <div className="grid flex-1 grid-cols-3 items-center gap-5 sm:grid-cols-5">
            {integrations.map((name) => (
              <span
                key={name}
                className="text-center font-heading text-sm font-bold tracking-[0.14em] text-white/25"
              >
                {name}
              </span>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};
