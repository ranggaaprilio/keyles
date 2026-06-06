import { Link } from "react-router-dom";
import { isAuthenticated } from "@/services/api/auth";
import { Menu, Search, User } from "lucide-react";

export const Navbar = () => {
  const isLoggedIn = isAuthenticated();

  return (
    <header className="sticky left-0 right-0 top-0 z-50 bg-black">
      {/* Top banner — Dell 1996 style */}
      <div className="mx-auto max-w-[760px] px-2">
        <div className="flex items-center justify-between py-3 px-4">
          {/* Left: Brand */}
          <Link to="/" className="flex items-center gap-3" aria-label="Keyles">
            <span className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-xl font-black uppercase text-white">
              KEYLES
            </span>
            <span className="hidden font-['Times_New_Roman',Times,serif] text-sm text-gray-300 sm:inline">
              .com
            </span>
          </Link>

          {/* Center: Headline */}
          <div className="hidden text-center md:block">
            <p className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-white">
              BUILD YOUR OWN SSO. ONLINE.
            </p>
          </div>

          {/* Right: Phone + BUY sticker */}
          <div className="flex items-center gap-3">
            {/* Phone callout */}
            <a
              href="tel:1-800-213-DELL"
              className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold text-[#e91d2a]"
            >
              1-800-KEYLES
            </a>

            {/* BUY a DELL sticker */}
            <div className="border border-black bg-[#fcc20f] px-2 py-0.5">
              <span className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-black">
                START
              </span>
              <span className="ml-1 inline-block bg-[#6a26a4] px-1 font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold text-white">
                a
              </span>
              <span className="ml-1 font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-black">
                KEYLES
              </span>
            </div>

            {/* Mobile menu */}
            <button
              className="inline-flex h-8 w-8 items-center justify-center text-white md:hidden"
              aria-label="Open menu"
            >
              <Menu className="h-5 w-5" />
            </button>
          </div>
        </div>

        {/* Nav links row */}
        <div className="flex items-center justify-between border-t border-gray-700 py-1.5 px-4">
          <nav className="flex items-center gap-4">
            {isLoggedIn ? (
              <Link
                to="/dashboard"
                className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-[#0000ee] underline hover:text-[#551a8b]"
              >
                Dashboard
              </Link>
            ) : (
              <>
                <Link
                  to="/login"
                  className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-[#0000ee] underline hover:text-[#551a8b]"
                >
                  Login
                </Link>
                <Link
                  to="/register"
                  className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-[#0000ee] underline hover:text-[#551a8b]"
                >
                  Register
                </Link>
              </>
            )}
            <Link
              to="/#how-it-works"
              className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-[#0000ee] underline hover:text-[#551a8b]"
            >
              Platform
            </Link>
            <Link
              to="/#about"
              className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-[#0000ee] underline hover:text-[#551a8b]"
            >
              Security
            </Link>
            <Link
              to="/#pricing"
              className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-[#0000ee] underline hover:text-[#551a8b]"
            >
              Plans
            </Link>
            <Link
              to="/docs/oauth"
              className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-[#0000ee] underline hover:text-[#551a8b]"
            >
              Docs
            </Link>
          </nav>
          <div className="flex items-center gap-2">
            <button
              className="inline-flex h-6 w-6 items-center justify-center text-gray-400 hover:text-white"
              aria-label="Search"
            >
              <Search className="h-3.5 w-3.5" />
            </button>
            <button
              className="inline-flex h-6 w-6 items-center justify-center text-gray-400 hover:text-white"
              aria-label="Account"
            >
              <User className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      </div>
    </header>
  );
};
