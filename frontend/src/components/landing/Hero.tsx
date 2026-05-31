import { Button } from "@/components/ui/button";
import { ScrollReveal } from "./ScrollReveal";
import { Link } from "react-router-dom";
import { isAuthenticated } from "@/services/api/auth";

export const Hero = () => {
  const isLoggedIn = isAuthenticated();

  return (
    <section className="bg-white">
      <div className="mx-auto max-w-[760px]">
        {/* Section eyebrow — olive tint */}
        <div className="bg-[#8e8a25] px-4 py-6">
          <ScrollReveal width="100%">
            <h1 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[36px] font-black uppercase leading-[1.0] text-black">
              SINGLE SIGN-ON<br />
              PLATFORM
            </h1>
          </ScrollReveal>
        </div>

        {/* Ribbon card body — CTA red panel */}
        <div className="border-x border-b border-black bg-[#e91d2a] p-4">
          <ScrollReveal width="100%" delay={0.1}>
            <p className="font-['Times_New_Roman',Times,serif] text-sm leading-[1.4] text-[#fffff0]">
              At Keyles, we&apos;ll help you configure SSO, manage clients,
              enforce session policies, and map roles — all from one
              operating surface built for speed and control.
            </p>
          </ScrollReveal>
        </div>

        {/* Product ribbon cards */}
        <div className="mt-0">
          {/* One Identity — sage */}
          <ScrollReveal width="100%" delay={0.2}>
            <div className="border-x border-b border-black">
              <div className="border-b border-black bg-white px-3 py-1.5">
                <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                  ONE IDENTITY
                </h3>
              </div>
              <div className="bg-[#b3bd95] px-4 py-3">
                <p className="font-['Times_New_Roman',Times,serif] text-sm leading-[1.4] text-black">
                  Users authenticate once, then move across approved
                  applications without password repetition. SSO keeps the
                  surface clean and the identity chain tight.
                </p>
              </div>
            </div>
          </ScrollReveal>

          {/* Policy Enforced — salmon */}
          <ScrollReveal width="100%" delay={0.3}>
            <div className="border-x border-b border-black">
              <div className="border-b border-black bg-white px-3 py-1.5">
                <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                  POLICY ENFORCED
                </h3>
              </div>
              <div className="bg-[#d77a7a] px-4 py-3">
                <p className="font-['Times_New_Roman',Times,serif] text-sm leading-[1.4] text-black">
                  Tenant settings, OAuth clients, and session controls stay
                  visible to administrators. Every policy change is tracked,
                  every boundary is enforced.
                </p>
              </div>
            </div>
          </ScrollReveal>

          {/* Access Mapped — sky */}
          <ScrollReveal width="100%" delay={0.4}>
            <div className="border-x border-b border-black">
              <div className="border-b border-black bg-white px-3 py-1.5">
                <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                  ACCESS MAPPED
                </h3>
              </div>
              <div className="bg-[#9ab6c8] px-4 py-3">
                <p className="font-['Times_New_Roman',Times,serif] text-sm leading-[1.4] text-black">
                  Roles and client grants keep each application tied to the
                  right user boundary. One surface, complete visibility.
                </p>
              </div>
            </div>
          </ScrollReveal>
        </div>

        {/* CTAs */}
        <ScrollReveal width="100%" delay={0.5}>
          <div className="flex gap-2 border-x border-b border-black bg-white p-4">
            <Button asChild>
              <Link to={isLoggedIn ? "/dashboard" : "/register"}>
                {isLoggedIn ? "Go to Dashboard" : "Start Now"}
              </Link>
            </Button>
            <Button variant="outline" asChild>
              <Link to="/#how-it-works">View Platform</Link>
            </Button>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
};
