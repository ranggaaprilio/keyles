import { Navbar } from "../components/landing/Navbar";
import { Hero } from "../components/landing/Hero";
import { SSOExplanation } from "../components/landing/SSOExplanation";
import { About } from "../components/landing/About";
import { Pricing } from "../components/landing/Pricing";
import { Footer } from "../components/landing/Footer";

const LandingPage = () => {
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <main className="flex-grow">
        <Hero />
        <SSOExplanation />
        <About />
        <Pricing />
      </main>
      <Footer />
    </div>
  );
};

export default LandingPage;
