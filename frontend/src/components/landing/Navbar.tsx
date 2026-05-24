import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { isAuthenticated } from "@/services/api/auth";
import { Menu, Search, User } from "lucide-react";

export const Navbar = () => {
  const isLoggedIn = isAuthenticated();

  return (
    <header className="fixed left-0 right-0 top-0 z-50 border-b border-[#3c3c3c] bg-black">
      <div className="mx-auto flex h-16 max-w-[1440px] items-center justify-between px-5 md:px-8">
        <Link to="/" className="flex items-center gap-3" aria-label="Keyles">
          <span className="grid h-8 w-8 grid-cols-3">
            <span className="bg-[#0066b1]" />
            <span className="bg-[#1c69d4]" />
            <span className="bg-[#e22718]" />
          </span>
          <span className="text-xl font-bold uppercase tracking-[1.5px] text-white">
            Keyles
          </span>
        </Link>

        <nav className="hidden items-center gap-8 md:flex">
          <Link
            to="/#how-it-works"
            className="text-sm font-light uppercase tracking-[1.5px] text-[#bbbbbb] transition-colors hover:text-white"
          >
            Platform
          </Link>
          <Link
            to="/#about"
            className="text-sm font-light uppercase tracking-[1.5px] text-[#bbbbbb] transition-colors hover:text-white"
          >
            Security
          </Link>
          <Link
            to="/#pricing"
            className="text-sm font-light uppercase tracking-[1.5px] text-[#bbbbbb] transition-colors hover:text-white"
          >
            Plans
          </Link>
        </nav>

        <div className="flex items-center gap-3">
          <button
            className="hidden h-12 w-12 items-center justify-center rounded-full bg-[#1a1a1a] text-white md:inline-flex"
            aria-label="Search"
          >
            <Search className="h-5 w-5" />
          </button>
          <button
            className="hidden h-12 w-12 items-center justify-center rounded-full bg-[#1a1a1a] text-white md:inline-flex"
            aria-label="Account"
          >
            <User className="h-5 w-5" />
          </button>
          {isLoggedIn ? (
            <Button asChild className="hidden md:inline-flex">
              <Link to="/dashboard">Dashboard</Link>
            </Button>
          ) : (
            <Button asChild className="hidden md:inline-flex">
              <Link to="/login">Login</Link>
            </Button>
          )}
          <button
            className="inline-flex h-12 w-12 items-center justify-center rounded-full bg-[#1a1a1a] text-white md:hidden"
            aria-label="Open menu"
          >
            <Menu className="h-5 w-5" />
          </button>
        </div>
      </div>
    </header>
  );
};
