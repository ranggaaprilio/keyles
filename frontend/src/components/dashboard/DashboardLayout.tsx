/**
 * Dashboard Layout Component
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
  BookOpen,
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
  {
    label: "Integration Guide",
    href: "/dashboard/integration",
    icon: <BookOpen className="w-4 h-4" />,
    matchPaths: ["/dashboard/integration"],
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
    <div className="min-h-screen bg-background flex">
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 w-60 bg-muted border-r border-border flex flex-col transition-transform duration-base lg:translate-x-0 lg:static lg:z-auto",
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        {/* Sidebar header */}
        <div className="flex items-center gap-3 px-4 py-4">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground">
            <KeyRound className="w-4 h-4" />
          </span>
          <div className="min-w-0">
            <h1 className="text-sm font-semibold text-foreground truncate">
              Keyles
            </h1>
            <p className="text-xs text-muted-foreground truncate">
              {tenant?.organization_name ?? "SSO Platform"}
            </p>
          </div>
          <button
            onClick={() => setSidebarOpen(false)}
            className="ml-auto lg:hidden rounded-md p-1 text-muted-foreground hover:text-foreground"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Nav items */}
        <nav className="flex-1 px-2 py-2 overflow-y-auto">
          {navItems.map((item) => {
            const active = isNavActive(item, location.pathname);
            return (
              <Link
                key={item.href}
                to={item.href}
                onClick={() => setSidebarOpen(false)}
                className={cn(
                  "flex items-center gap-2 my-0.5 rounded-md px-3 py-2 text-sm font-medium transition-colors duration-fast",
                  active
                    ? "bg-background text-foreground shadow-card"
                    : "text-muted-foreground hover:bg-background hover:text-foreground"
                )}
              >
                <span className={cn(active ? "text-foreground" : "text-muted-foreground")}>
                  {item.icon}
                </span>
                {item.label}
              </Link>
            );
          })}
        </nav>

        {/* Sidebar footer */}
        <div className="border-t border-border px-4 py-3">
          {user && (
            <div className="mb-2">
              <p className="text-sm font-medium text-foreground truncate">
                {user.full_name}
              </p>
              <p className="text-xs text-muted-foreground truncate">
                {user.email}
              </p>
            </div>
          )}
          <button
            onClick={logout}
            className="flex items-center gap-2 w-full rounded-md py-1.5 text-sm text-muted-foreground hover:text-destructive transition-colors duration-fast"
          >
            <LogOut className="w-4 h-4" />
            Logout
          </button>
        </div>
      </aside>

      <div className="flex-1 flex flex-col min-w-0">
        <header className="sticky top-0 z-30 h-14 bg-background/90 backdrop-blur border-b border-border px-4 sm:px-6 flex items-center gap-3">
          <button
            onClick={() => setSidebarOpen(true)}
            className="lg:hidden rounded-md p-1 -ml-1 text-foreground"
            aria-label="Open sidebar"
          >
            <Menu className="w-5 h-5" />
          </button>

          <PageTitle pathname={location.pathname} />

          <div className="ml-auto hidden sm:flex items-center gap-2">
            {tenant && (
              <span className="inline-flex items-center rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-foreground">
                {tenant.organization_name}
              </span>
            )}
          </div>
        </header>

        <main className="flex-1 overflow-y-auto bg-background">{children}</main>
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
  } else if (pathname === "/dashboard/integration") {
    title = "Integration Guide";
  }

  return (
    <div className="flex items-center gap-2">
      {backHref && (
        <Link
          to={backHref}
          className="rounded-md p-1 text-muted-foreground hover:text-foreground"
        >
          <ChevronLeft className="w-5 h-5" />
        </Link>
      )}
      <h2 className="text-base font-medium text-foreground">
        {title}
      </h2>
    </div>
  );
}
