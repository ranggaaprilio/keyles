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
    <footer className="bg-black text-white">
      <div className="mx-auto max-w-7xl px-5 py-14 sm:px-8 lg:px-12">
        <div className="grid gap-10 border-b border-white/10 pb-12 md:grid-cols-[1fr_auto] md:items-start">
          <div>
            <Link
              to="/"
              className="inline-flex items-center gap-3 text-white no-underline"
            >
              <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-[#e91d2a]">
                <ShieldCheck className="h-5 w-5" />
              </span>
              <span className="font-heading text-xl font-bold tracking-[-0.04em]">
                Keyles
              </span>
            </Link>
            <p className="mt-5 max-w-sm font-ui text-sm leading-6 text-white/45">
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
                className="inline-flex items-center gap-1 font-ui text-sm font-medium text-white/55 no-underline transition-colors hover:text-white"
              >
                {item.label}
                {item.label === "Documentation" && (
                  <ArrowUpRight className="h-3.5 w-3.5" />
                )}
              </Link>
            ))}
          </nav>
        </div>

        <div className="flex flex-col gap-4 pt-7 font-ui text-xs text-white/30 sm:flex-row sm:items-center sm:justify-between">
          <p>&copy; {new Date().getFullYear()} Keyles. All rights reserved.</p>
          <div className="flex gap-6">
            <a href="#" className="text-white/30 no-underline hover:text-white">
              Privacy
            </a>
            <a href="#" className="text-white/30 no-underline hover:text-white">
              Terms
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
};
