import { Button } from "@/components/ui/button";
import { ScrollReveal } from "./ScrollReveal";
import { Link } from "react-router-dom";
import { isAuthenticated } from "@/services/api/auth";

export const Hero = () => {
  const isLoggedIn = isAuthenticated();

  return (
    <section className="min-h-screen flex items-center justify-center pt-16 bg-gradient-to-b from-background to-muted/20">
      <div className="container mx-auto px-4 text-center">
        <ScrollReveal width="100%">
          <h1 className="text-4xl md:text-6xl font-bold tracking-tight mb-6">
            Simplify Access with{" "}
            <span className="text-primary">Single Sign-On</span>
          </h1>
        </ScrollReveal>

        <ScrollReveal width="100%" delay={0.2}>
          <p className="text-xl text-muted-foreground mb-8 max-w-2xl mx-auto">
            Secure, seamless, and scalable authentication for your enterprise.
            Manage all your applications with one identity.
          </p>
        </ScrollReveal>

        <ScrollReveal width="100%" delay={0.4}>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button size="lg" asChild>
              <Link to={isLoggedIn ? "/dashboard" : "/register"}>
                {isLoggedIn ? "Go to Dashboard" : "Get Started"}
              </Link>
            </Button>
            <Button size="lg" variant="outline" asChild>
              <Link to="/#about">Learn More</Link>
            </Button>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
};
