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
    <section className="relative overflow-hidden bg-background text-foreground">
      <div className="relative mx-auto grid max-w-7xl items-center gap-16 px-5 pb-20 pt-16 sm:px-8 sm:pb-28 sm:pt-24 lg:grid-cols-[0.9fr_1.1fr] lg:px-12 lg:py-32">
        <ScrollReveal width="100%">
          <div>
            <div className="mb-7 inline-flex items-center gap-2 rounded-full border border-border bg-muted px-4 py-1.5 text-xs font-medium text-muted-foreground">
              <span className="h-2 w-2 rounded-full bg-success" />
              Identity infrastructure, simplified
            </div>
            <h1 className="max-w-2xl text-5xl font-normal leading-[1.02] tracking-[-0.04em] text-foreground sm:text-6xl lg:text-7xl">
              One login.
              <br />
              Every app.
              <br />
              <span className="text-muted-foreground">Total control.</span>
            </h1>
            <p className="mt-7 max-w-xl text-base leading-7 text-muted-foreground sm:text-lg">
              Give every team secure access to the tools they need. Keyles
              brings SSO, roles, sessions, and OAuth clients into one clear
              operating surface.
            </p>
            <div className="mt-9 flex flex-col gap-3 sm:flex-row">
              <Link
                to={isLoggedIn ? "/dashboard" : "/register"}
                className="inline-flex h-12 items-center justify-center gap-2 rounded-full bg-primary px-7 text-sm font-medium text-primary-foreground no-underline transition-colors duration-fast hover:bg-primary/85 hover:text-primary-foreground"
              >
                {isLoggedIn ? "Open dashboard" : "Start for free"}
                <ArrowRight className="h-4 w-4" />
              </Link>
              <Link
                to="/docs/oauth"
                className="inline-flex h-12 items-center justify-center rounded-full border border-input bg-background px-7 text-sm font-medium text-foreground no-underline transition-colors duration-fast hover:bg-muted hover:text-foreground"
              >
                Explore documentation
              </Link>
            </div>
            <div className="mt-8 flex flex-wrap gap-x-6 gap-y-3 text-xs text-muted-foreground">
              {["No credit card", "Deploy in minutes", "OAuth 2.0 ready"].map(
                (item) => (
                  <span key={item} className="flex items-center gap-2">
                    <Check className="h-3.5 w-3.5 text-foreground" />
                    {item}
                  </span>
                ),
              )}
            </div>
          </div>
        </ScrollReveal>

        <ScrollReveal width="100%" delay={0.15}>
          <div className="relative mx-auto w-full max-w-2xl lg:mx-0">
            <div className="relative overflow-hidden rounded-xl border border-border bg-card shadow-elevated">
              <div className="flex items-center justify-between border-b border-border px-5 py-3.5">
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full bg-border" />
                  <span className="h-2.5 w-2.5 rounded-full bg-border" />
                  <span className="h-2.5 w-2.5 rounded-full bg-border" />
                </div>
                <span className="text-xs text-muted-foreground">
                  console.keyles.io
                </span>
                <MoreHorizontal className="h-4 w-4 text-muted-foreground" />
              </div>

              <div className="grid min-h-[430px] grid-cols-[72px_1fr] sm:grid-cols-[176px_1fr]">
                <aside className="border-r border-border bg-muted p-3 sm:p-4">
                  <div className="mb-7 flex items-center gap-2 px-2">
                    <span className="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-primary-foreground">
                      <Fingerprint className="h-4 w-4" />
                    </span>
                    <span className="hidden text-sm font-medium sm:block">
                      Keyles
                    </span>
                  </div>
                  <div className="space-y-1">
                    {[
                      { label: "Overview", icon: ShieldCheck, active: true },
                      { label: "Users", icon: Users },
                      { label: "Clients", icon: Fingerprint },
                    ].map((item) => (
                      <div
                        key={item.label}
                        className={`flex items-center gap-3 rounded-md px-2.5 py-2 text-sm ${
                          item.active
                            ? "bg-background text-foreground shadow-card"
                            : "text-muted-foreground"
                        }`}
                      >
                        <item.icon className="h-4 w-4 shrink-0" />
                        <span className="hidden text-xs font-medium sm:block">
                          {item.label}
                        </span>
                      </div>
                    ))}
                  </div>
                </aside>

                <div className="min-w-0 p-4 sm:p-6">
                  <div className="mb-6 flex items-center justify-between">
                    <div>
                      <p className="text-xs text-muted-foreground">Overview</p>
                      <h2 className="mt-1 text-lg font-medium tracking-tight sm:text-xl">
                        Good morning, Alex
                      </h2>
                    </div>
                    <button
                      type="button"
                      className="flex items-center gap-2 rounded-md border border-border px-3 py-1.5 text-xs text-muted-foreground"
                    >
                      ACME Inc.
                      <ChevronDown className="h-3 w-3" />
                    </button>
                  </div>

                  <div className="grid gap-3 sm:grid-cols-2">
                    <div className="rounded-lg border border-border p-4">
                      <div className="flex items-start justify-between">
                        <span className="text-xs text-muted-foreground">
                          Active users
                        </span>
                        <Users className="h-4 w-4 text-muted-foreground" />
                      </div>
                      <p className="mt-5 text-3xl font-normal tracking-tight">2,418</p>
                      <p className="mt-1 text-[11px] text-success">
                        +12.4% this month
                      </p>
                    </div>
                    <div className="rounded-lg border border-border p-4">
                      <div className="flex items-start justify-between">
                        <span className="text-xs text-muted-foreground">
                          Success rate
                        </span>
                        <ShieldCheck className="h-4 w-4 text-muted-foreground" />
                      </div>
                      <p className="mt-5 text-3xl font-normal tracking-tight">
                        99.98%
                      </p>
                      <p className="mt-1 text-[11px] text-muted-foreground">
                        Last 30 days
                      </p>
                    </div>
                  </div>

                  <div className="mt-3 rounded-lg border border-border p-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-xs font-medium">
                          Authentication activity
                        </p>
                        <p className="mt-1 text-[11px] text-muted-foreground">
                          Successful sign-ins
                        </p>
                      </div>
                      <span className="text-xs text-muted-foreground">7 days</span>
                    </div>
                    <div className="mt-6 flex h-24 items-end gap-2">
                      {[40, 65, 48, 78, 60, 90, 72, 98, 84, 105, 94, 118].map(
                        (height, index) => (
                          <div
                            key={`${height}-${index}`}
                            className="flex-1 rounded-t-sm bg-primary/80"
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

      <div className="relative border-t border-border">
        <div className="mx-auto flex max-w-7xl flex-col gap-6 px-5 py-8 sm:px-8 lg:flex-row lg:items-center lg:px-12">
          <p className="shrink-0 text-xs font-medium text-muted-foreground">
            Built for modern teams
          </p>
          <div className="grid flex-1 grid-cols-3 items-center gap-5 sm:grid-cols-5">
            {integrations.map((name) => (
              <span
                key={name}
                className="text-center text-sm font-medium tracking-[0.14em] text-muted-foreground/60"
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
