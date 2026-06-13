import { useState } from "react";
import { Link } from "react-router-dom";
import { ArrowUpRight, Menu, ShieldCheck, X } from "lucide-react";
import { isAuthenticated } from "@/services/api/auth";

const navigation = [
  { label: "Platform", href: "/#how-it-works" },
  { label: "Security", href: "/#about" },
  { label: "Pricing", href: "/#pricing" },
  { label: "Docs", href: "/docs/oauth" },
];

export const Navbar = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const isLoggedIn = isAuthenticated();

  return (
    <header className="sticky top-0 z-50 border-b border-white/10 bg-black/90 text-white backdrop-blur-xl">
      <div className="mx-auto flex h-20 max-w-7xl items-center justify-between px-5 sm:px-8 lg:px-12">
        <Link
          to="/"
          className="flex items-center gap-3 text-white no-underline"
          aria-label="Keyles home"
        >
          <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-[#e91d2a] shadow-[0_8px_30px_rgba(233,29,42,0.24)]">
            <ShieldCheck className="h-5 w-5" />
          </span>
          <span className="font-heading text-xl font-bold tracking-[-0.04em]">
            Keyles
          </span>
        </Link>

        <nav className="hidden items-center gap-8 lg:flex" aria-label="Primary">
          {navigation.map((item) => (
            <Link
              key={item.label}
              to={item.href}
              className="font-ui text-sm font-medium text-white/65 no-underline transition-colors hover:text-white"
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <div className="hidden items-center gap-3 lg:flex">
          {!isLoggedIn && (
            <Link
              to="/login"
              className="px-4 py-2 font-ui text-sm font-semibold text-white no-underline transition-colors hover:text-white/70"
            >
              Sign in
            </Link>
          )}
          <Link
            to={isLoggedIn ? "/dashboard" : "/register"}
            className="inline-flex h-11 items-center gap-2 rounded-xl bg-[#e91d2a] px-5 font-ui text-sm font-bold text-white no-underline shadow-[0_8px_30px_rgba(233,29,42,0.2)] transition-all hover:-translate-y-0.5 hover:bg-[#cf1723] hover:text-white"
          >
            {isLoggedIn ? "Dashboard" : "Get started"}
            <ArrowUpRight className="h-4 w-4" />
          </Link>
        </div>

        <button
          type="button"
          className="inline-flex h-11 w-11 items-center justify-center rounded-xl border border-white/15 text-white lg:hidden"
          aria-label={isMenuOpen ? "Close menu" : "Open menu"}
          aria-expanded={isMenuOpen}
          onClick={() => setIsMenuOpen((open) => !open)}
        >
          {isMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {isMenuOpen && (
        <div className="border-t border-white/10 bg-black px-5 py-5 lg:hidden">
          <nav className="mx-auto flex max-w-7xl flex-col" aria-label="Mobile">
            {navigation.map((item) => (
              <Link
                key={item.label}
                to={item.href}
                onClick={() => setIsMenuOpen(false)}
                className="border-b border-white/10 py-4 font-ui text-base font-medium text-white no-underline"
              >
                {item.label}
              </Link>
            ))}
            <div className="mt-5 grid grid-cols-2 gap-3">
              {!isLoggedIn && (
                <Link
                  to="/login"
                  className="inline-flex h-12 items-center justify-center rounded-xl border border-white/20 font-ui font-bold text-white no-underline"
                >
                  Sign in
                </Link>
              )}
              <Link
                to={isLoggedIn ? "/dashboard" : "/register"}
                className="inline-flex h-12 items-center justify-center rounded-xl bg-[#e91d2a] font-ui font-bold text-white no-underline"
              >
                {isLoggedIn ? "Dashboard" : "Get started"}
              </Link>
            </div>
          </nav>
        </div>
      )}
    </header>
  );
};
