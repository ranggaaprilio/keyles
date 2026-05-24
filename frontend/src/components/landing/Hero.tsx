import { Button } from "@/components/ui/button";
import { ScrollReveal } from "./ScrollReveal";
import { Link } from "react-router-dom";
import { isAuthenticated } from "@/services/api/auth";

export const Hero = () => {
  const isLoggedIn = isAuthenticated();

  return (
    <section className="relative min-h-[92vh] overflow-hidden bg-black pt-16">
      <div
        className="absolute inset-0 bg-cover bg-center"
        style={{
          backgroundImage:
            "url('https://images.unsplash.com/photo-1492144534655-ae79c964c9d7?auto=format&fit=crop&w=2400&q=85')",
        }}
        aria-hidden="true"
      />
      <div className="absolute inset-0 bg-black/55" aria-hidden="true" />
      <div
        className="absolute bottom-0 left-0 right-0 h-36 bg-gradient-to-t from-black to-transparent"
        aria-hidden="true"
      />

      <div className="relative mx-auto flex min-h-[calc(92vh-4rem)] max-w-[1440px] items-end px-5 pb-16 pt-24 md:px-8 md:pb-20">
        <div className="max-w-4xl">
          <ScrollReveal width="100%">
            <div className="mb-8 grid h-1 w-56 grid-cols-3">
              <span className="bg-[#0066b1]" />
              <span className="bg-[#1c69d4]" />
              <span className="bg-[#e22718]" />
            </div>
            <p className="mb-4 text-sm font-bold uppercase tracking-[1.5px] text-white">
              Identity engineered for speed
            </p>
            <h1 className="text-5xl font-bold uppercase leading-none text-white md:text-7xl lg:text-[80px]">
              Single sign-on with full control.
            </h1>
          </ScrollReveal>

          <ScrollReveal width="100%" delay={0.2}>
            <p className="mt-6 max-w-2xl text-lg font-light leading-8 text-[#e6e6e6] md:text-xl">
              Keyles gives every tenant a precise OAuth and OpenID Connect
              command center for users, clients, sessions, and access policy.
            </p>
          </ScrollReveal>

          <ScrollReveal width="100%" delay={0.4}>
            <div className="mt-10 flex flex-col gap-4 sm:flex-row">
              <Button size="lg" asChild>
                <Link to={isLoggedIn ? "/dashboard" : "/register"}>
                  {isLoggedIn ? "Go to dashboard" : "Start now"}
                </Link>
              </Button>
              <Button size="lg" variant="outline" asChild>
                <Link to="/#how-it-works">View platform</Link>
              </Button>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>
  );
};
