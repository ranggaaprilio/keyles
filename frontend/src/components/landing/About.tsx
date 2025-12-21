import { ScrollReveal } from "./ScrollReveal";
import { Card, CardContent } from "@/components/ui/card";

export const About = () => {
  return (
    <section id="about" className="py-20 bg-muted/30">
      <div className="container mx-auto px-4">
        <ScrollReveal width="100%">
          <h2 className="text-3xl font-bold text-center mb-12">About Keyles</h2>
        </ScrollReveal>

        <div className="max-w-4xl mx-auto">
          <ScrollReveal width="100%" delay={0.2}>
            <Card className="bg-background/50 backdrop-blur-sm border-none shadow-sm">
              <CardContent className="p-8 text-center space-y-6">
                <p className="text-lg text-muted-foreground leading-relaxed">
                  Keyles is dedicated to simplifying identity management for
                  modern enterprises. We believe that security shouldn't come at
                  the cost of user experience.
                </p>
                <p className="text-lg text-muted-foreground leading-relaxed">
                  Our mission is to provide a seamless, secure, and scalable
                  Single Sign-On solution that empowers your team to focus on
                  what matters most—building great products.
                </p>
              </CardContent>
            </Card>
          </ScrollReveal>
        </div>
      </div>
    </section>
  );
};
