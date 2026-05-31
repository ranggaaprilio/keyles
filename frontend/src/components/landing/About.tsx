import { ScrollReveal } from "./ScrollReveal";

export const About = () => {
  return (
    <section id="about" className="bg-white">
      <div className="mx-auto max-w-[760px]">
        {/* Section eyebrow — salmon */}
        <ScrollReveal width="100%">
          <div className="bg-[#d77a7a] px-4 py-6">
            <p className="mb-1 font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
              Security Control
            </p>
            <h2 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[36px] font-black uppercase leading-[1.0] text-black">
              BUILT FOR TENANTS<br />
              THAT CANNOT LOSE<br />
              THE LINE.
            </h2>
          </div>
        </ScrollReveal>

        {/* CTA red panel */}
        <ScrollReveal width="100%" delay={0.15}>
          <div className="border-x border-b border-black bg-[#e91d2a] p-4">
            <p className="font-['Times_New_Roman',Times,serif] text-sm leading-[1.4] text-[#fffff0]">
              Keyles keeps identity management direct: tenant onboarding,
              OAuth clients, session visibility, and user roles all share one
              operating surface. Security stays strict without burying teams in
              avoidable ceremony.
            </p>
          </div>
        </ScrollReveal>

        {/* Support ribbon card — steel */}
        <ScrollReveal width="100%" delay={0.3}>
          <div className="border-x border-b border-black">
            <div className="border-b border-black bg-white px-3 py-1.5">
              <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                FROM KEYLES&apos; AWARD-WINNING SERVICE AND SUPPORT TEAMS
              </h3>
            </div>
            <div className="bg-[#a5b8c0] px-4 py-3">
              <p className="font-['Times_New_Roman',Times,serif] text-sm leading-[1.4] text-black">
                Every tenant comes with built-in monitoring, session audit
                trails, and direct configuration tools. No third-party
                dependencies, no hidden complexity — just clear identity
                infrastructure that works.
              </p>
            </div>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
};
