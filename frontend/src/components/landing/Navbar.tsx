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
    <header className="sticky top-0 z-50 border-b border-border bg-background/90 text-foreground backdrop-blur">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-5 sm:px-8 lg:px-12">
        <Link
          to="/"
          className="flex items-center gap-2.5 text-foreground no-underline"
          aria-label="Keyles home"
        >
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary text-primary-foreground">
            <ShieldCheck className="h-4 w-4" />
          </span>
          <span className="text-lg font-medium tracking-tight">Keyles</span>
        </Link>

        <nav className="hidden items-center gap-8 lg:flex" aria-label="Primary">
          {navigation.map((item) => (
            <Link
              key={item.label}
              to={item.href}
              className="text-sm font-medium text-muted-foreground no-underline transition-colors duration-fast hover:text-foreground"
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <div className="hidden items-center gap-2 lg:flex">
          {!isLoggedIn && (
            <Link
              to="/login"
              className="px-4 py-2 text-sm font-medium text-foreground no-underline transition-colors duration-fast hover:text-muted-foreground"
            >
              Sign in
            </Link>
          )}
          <Link
            to={isLoggedIn ? "/dashboard" : "/register"}
            className="inline-flex h-10 items-center gap-2 rounded-full bg-primary px-5 text-sm font-medium text-primary-foreground no-underline transition-colors duration-fast hover:bg-primary/85 hover:text-primary-foreground"
          >
            {isLoggedIn ? "Dashboard" : "Get started"}
            <ArrowUpRight className="h-4 w-4" />
          </Link>
        </div>

        <button
          type="button"
          className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-border text-foreground lg:hidden"
          aria-label={isMenuOpen ? "Close menu" : "Open menu"}
          aria-expanded={isMenuOpen}
          onClick={() => setIsMenuOpen((open) => !open)}
        >
          {isMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {isMenuOpen && (
        <div className="border-t border-border bg-background px-5 py-5 lg:hidden">
          <nav className="mx-auto flex max-w-7xl flex-col" aria-label="Mobile">
            {navigation.map((item) => (
              <Link
                key={item.label}
                to={item.href}
                onClick={() => setIsMenuOpen(false)}
                className="border-b border-border py-4 text-base font-medium text-foreground no-underline"
              >
                {item.label}
              </Link>
            ))}
            <div className="mt-5 grid grid-cols-2 gap-3">
              {!isLoggedIn && (
                <Link
                  to="/login"
                  className="inline-flex h-11 items-center justify-center rounded-full border border-input text-sm font-medium text-foreground no-underline hover:bg-muted"
                >
                  Sign in
                </Link>
              )}
              <Link
                to={isLoggedIn ? "/dashboard" : "/register"}
                className="inline-flex h-11 items-center justify-center rounded-full bg-primary text-sm font-medium text-primary-foreground no-underline hover:bg-primary/85 hover:text-primary-foreground"
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
