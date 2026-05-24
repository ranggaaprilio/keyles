import { ScrollReveal } from "./ScrollReveal";

export const About = () => {
  return (
    <section id="about" className="bg-black">
      <div
        className="min-h-[560px] bg-cover bg-center"
        style={{
          backgroundImage:
            "linear-gradient(90deg, rgba(0,0,0,.88), rgba(0,0,0,.28)), url('https://images.unsplash.com/photo-1503376780353-7e6692767b70?auto=format&fit=crop&w=2400&q=85')",
        }}
      >
        <div className="mx-auto flex min-h-[560px] max-w-[1440px] items-center px-5 py-24 md:px-8">
          <div className="max-w-2xl">
            <ScrollReveal width="100%">
              <p className="mb-3 text-sm font-bold uppercase tracking-[1.5px] text-[#bbbbbb]">
                Security control
              </p>
              <h2 className="text-4xl font-bold uppercase leading-tight text-white md:text-6xl">
                Built for tenants that cannot lose the line.
              </h2>
            </ScrollReveal>

            <ScrollReveal width="100%" delay={0.2}>
              <p className="mt-6 text-lg font-light leading-8 text-[#e6e6e6]">
                Keyles keeps identity management direct: tenant onboarding,
                OAuth clients, session visibility, and user roles all share one
                operating surface. Security stays strict without burying teams in
                avoidable ceremony.
              </p>
            </ScrollReveal>
          </div>
        </div>
      </div>
    </section>
  );
};
