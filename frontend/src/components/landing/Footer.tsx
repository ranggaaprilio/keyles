import { ArrowUpRight, ShieldCheck } from "lucide-react";
import { Link } from "react-router-dom";

const links = [
  { label: "Platform", href: "/#how-it-works" },
  { label: "Security", href: "/#about" },
  { label: "Pricing", href: "/#pricing" },
  { label: "Documentation", href: "/docs/oauth" },
];

export const Footer = () => {
  return (
    <footer className="border-t border-border bg-background text-foreground">
      <div className="mx-auto max-w-7xl px-5 py-14 sm:px-8 lg:px-12">
        <div className="grid gap-10 border-b border-border pb-12 md:grid-cols-[1fr_auto] md:items-start">
          <div>
            <Link
              to="/"
              className="inline-flex items-center gap-2.5 text-foreground no-underline"
            >
              <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary text-primary-foreground">
                <ShieldCheck className="h-4 w-4" />
              </span>
              <span className="text-lg font-medium tracking-tight">Keyles</span>
            </Link>
            <p className="mt-5 max-w-sm text-sm leading-6 text-muted-foreground">
              Modern identity infrastructure for teams that want secure access
              without unnecessary complexity.
            </p>
          </div>

          <nav
            className="grid grid-cols-2 gap-x-10 gap-y-4 sm:grid-cols-4"
            aria-label="Footer"
          >
            {links.map((item) => (
              <Link
                key={item.label}
                to={item.href}
                className="inline-flex items-center gap-1 text-sm font-medium text-muted-foreground no-underline transition-colors duration-fast hover:text-foreground"
              >
                {item.label}
                {item.label === "Documentation" && (
                  <ArrowUpRight className="h-3.5 w-3.5" />
                )}
              </Link>
            ))}
          </nav>
        </div>

        <div className="flex flex-col gap-4 pt-7 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
          <p>&copy; {new Date().getFullYear()} Keyles. All rights reserved.</p>
          <div className="flex gap-6">
            <a href="#" className="no-underline transition-colors duration-fast hover:text-foreground">
              Privacy
            </a>
            <a href="#" className="no-underline transition-colors duration-fast hover:text-foreground">
              Terms
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
};
