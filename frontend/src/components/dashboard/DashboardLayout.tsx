/**
 * Dashboard Layout Component — Dell 1996 retro style
 */

import { ReactNode, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../../hooks/useAuth";
import { cn } from "../../lib/utils";
import {
  LayoutDashboard,
  AppWindow,
  Users,
  LogOut,
  Menu,
  X,
  ChevronLeft,
  KeyRound,
} from "lucide-react";

interface DashboardLayoutProps {
  children: ReactNode;
}

interface NavItem {
  label: string;
  href: string;
  icon: ReactNode;
  matchPaths: string[];
}

const navItems: NavItem[] = [
  {
    label: "Dashboard",
    href: "/dashboard",
    icon: <LayoutDashboard className="w-4 h-4" />,
    matchPaths: ["/dashboard"],
  },
  {
    label: "Client Applications",
    href: "/dashboard/clients",
    icon: <AppWindow className="w-4 h-4" />,
    matchPaths: ["/dashboard/clients"],
  },
  {
    label: "Users",
    href: "/dashboard/users",
    icon: <Users className="w-4 h-4" />,
    matchPaths: ["/dashboard/users"],
  },
];

function isNavActive(item: NavItem, pathname: string): boolean {
  if (item.href === "/dashboard") return pathname === "/dashboard";
  return item.matchPaths.some((p) => pathname.startsWith(p));
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
  const location = useLocation();
  const { dashboardQuery, logout } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const tenant = dashboardQuery.data?.tenant;
  const user = dashboardQuery.data?.user;

  return (
    <div className="min-h-screen bg-white">
      {/* Page frame — black border around everything */}
      <div className="border-[8px] border-black min-h-screen flex max-sm:border-[2px] md:border-[4px] lg:border-[8px]">
        {sidebarOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/50 lg:hidden"
            onClick={() => setSidebarOpen(false)}
          />
        )}

        <aside
          className={cn(
            "fixed inset-y-0 left-0 z-50 w-56 bg-white border-r-2 border-black flex flex-col transition-transform duration-200 lg:translate-x-0 lg:static lg:z-auto",
            sidebarOpen ? "translate-x-0" : "-translate-x-full"
          )}
        >
          {/* Sidebar header — black banner */}
          <div className="flex items-center gap-2 px-4 py-3 bg-black">
            <KeyRound className="w-4 h-4 text-white" />
            <div className="min-w-0">
              <h1 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-white truncate">
                Keyles
              </h1>
              <p className="font-['Times_New_Roman',Times,serif] text-[10px] text-gray-400 truncate">
                {tenant?.organization_name ?? "SSO Platform"}
              </p>
            </div>
            <button
              onClick={() => setSidebarOpen(false)}
              className="ml-auto lg:hidden text-gray-400 hover:text-white"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* Nav items */}
          <nav className="flex-1 py-2 overflow-y-auto">
            {navItems.map((item) => {
              const active = isNavActive(item, location.pathname);
              return (
                <Link
                  key={item.href}
                  to={item.href}
                  onClick={() => setSidebarOpen(false)}
                  className={cn(
                    "flex items-center gap-2 px-4 py-2 font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] border-b border-black transition-colors",
                    active
                      ? "bg-black text-white"
                      : "text-black hover:bg-gray-100"
                  )}
                >
                  <span className={cn(active ? "text-white" : "text-gray-500")}>
                    {item.icon}
                  </span>
                  {item.label}
                </Link>
              );
            })}
          </nav>

          {/* Sidebar footer */}
          <div className="border-t-2 border-black px-4 py-3">
            {user && (
              <div className="mb-2">
                <p className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase text-black truncate">
                  {user.full_name}
                </p>
                <p className="font-['Times_New_Roman',Times,serif] text-[11px] text-gray-500 truncate">
                  {user.email}
                </p>
              </div>
            )}
            <button
              onClick={logout}
              className="flex items-center gap-2 w-full font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black hover:bg-red-50 hover:text-red-700 py-1.5 transition-colors"
            >
              <LogOut className="w-4 h-4" />
              Logout
            </button>
          </div>
        </aside>

        <div className="flex-1 flex flex-col min-w-0">
          {/* Header — white with black underline */}
          <header className="sticky top-0 z-30 bg-white border-b-2 border-black px-4 sm:px-6 py-2 flex items-center gap-3">
            <button
              onClick={() => setSidebarOpen(true)}
              className="lg:hidden p-1 -ml-1 text-black"
              aria-label="Open sidebar"
            >
              <Menu className="w-5 h-5" />
            </button>

            <PageTitle pathname={location.pathname} />

            <div className="ml-auto hidden sm:flex items-center gap-2">
              {tenant && (
                <span className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] bg-gray-100 border border-black px-2 py-0.5">
                  {tenant.organization_name}
                </span>
              )}
            </div>
          </header>

          <main className="flex-1 overflow-y-auto bg-white">{children}</main>
        </div>
      </div>
    </div>
  );
}

function PageTitle({ pathname }: { pathname: string }) {
  let title = "Dashboard";
  let backHref: string | null = null;

  if (pathname === "/dashboard") {
    title = "Dashboard";
  } else if (pathname === "/dashboard/clients") {
    title = "Client Applications";
  } else if (pathname === "/dashboard/clients/new") {
    title = "Register New Client";
    backHref = "/dashboard/clients";
  } else if (pathname.match(/^\/dashboard\/clients\/[^/]+\/edit$/)) {
    title = "Edit Client";
    backHref = pathname.replace("/edit", "");
  } else if (pathname.match(/^\/dashboard\/clients\/[^/]+$/)) {
    title = "Client Details";
    backHref = "/dashboard/clients";
  } else if (pathname === "/dashboard/users") {
    title = "User Management";
  } else if (pathname.match(/^\/dashboard\/users\/[^/]+$/)) {
    title = "User Details";
    backHref = "/dashboard/users";
  }

  return (
    <div className="flex items-center gap-2">
      {backHref && (
        <Link
          to={backHref}
          className="p-1 text-black hover:text-gray-600"
        >
          <ChevronLeft className="w-5 h-5" />
        </Link>
      )}
      <h2 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase tracking-[1.5px] text-black">
        {title}
      </h2>
    </div>
  );
}
