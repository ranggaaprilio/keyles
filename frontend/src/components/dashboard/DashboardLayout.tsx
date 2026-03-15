/**
 * Dashboard Layout Component
 *
 * Provides a persistent sidebar navigation and header for all dashboard pages.
 * Wraps page content with consistent layout including:
 * - Collapsible sidebar with navigation links
 * - Top header with org name and user info
 * - Responsive mobile menu
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
    icon: <LayoutDashboard className="w-5 h-5" />,
    matchPaths: ["/dashboard"],
  },
  {
    label: "Client Applications",
    href: "/dashboard/clients",
    icon: <AppWindow className="w-5 h-5" />,
    matchPaths: ["/dashboard/clients"],
  },
  {
    label: "Users",
    href: "/dashboard/users",
    icon: <Users className="w-5 h-5" />,
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
    <div className="min-h-screen bg-gray-50 flex">
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 w-64 bg-white border-r border-gray-200 flex flex-col transition-transform duration-200 lg:translate-x-0 lg:static lg:z-auto",
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        <div className="flex items-center gap-3 px-5 py-5 border-b border-gray-200">
          <div className="w-9 h-9 bg-blue-600 rounded-lg flex items-center justify-center flex-shrink-0">
            <KeyRound className="w-5 h-5 text-white" />
          </div>
          <div className="min-w-0">
            <h1 className="text-base font-bold text-gray-900 truncate">Keyles</h1>
            <p className="text-xs text-gray-500 truncate">
              {tenant?.organization_name ?? "SSO Platform"}
            </p>
          </div>
          <button
            onClick={() => setSidebarOpen(false)}
            className="ml-auto lg:hidden p-1 rounded hover:bg-gray-100"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          {navItems.map((item) => {
            const active = isNavActive(item, location.pathname);
            return (
              <Link
                key={item.href}
                to={item.href}
                onClick={() => setSidebarOpen(false)}
                className={cn(
                  "flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors",
                  active
                    ? "bg-blue-50 text-blue-700"
                    : "text-gray-700 hover:bg-gray-100 hover:text-gray-900"
                )}
              >
                <span
                  className={cn(active ? "text-blue-600" : "text-gray-400")}
                >
                  {item.icon}
                </span>
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="border-t border-gray-200 px-3 py-4">
          {user && (
            <div className="px-3 mb-3">
              <p className="text-sm font-medium text-gray-900 truncate">{user.full_name}</p>
              <p className="text-xs text-gray-500 truncate">{user.email}</p>
            </div>
          )}
          <button
            onClick={logout}
            className="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-sm font-medium text-gray-700 hover:bg-red-50 hover:text-red-700 transition-colors"
          >
            <LogOut className="w-5 h-5 text-gray-400" />
            Logout
          </button>
        </div>
      </aside>

      <div className="flex-1 flex flex-col min-w-0">
        <header className="sticky top-0 z-30 bg-white border-b border-gray-200 px-4 sm:px-6 lg:px-8 py-3 flex items-center gap-4">
          <button
            onClick={() => setSidebarOpen(true)}
            className="lg:hidden p-2 -ml-2 rounded-lg hover:bg-gray-100"
            aria-label="Open sidebar"
          >
            <Menu className="w-5 h-5 text-gray-600" />
          </button>

          <PageTitle pathname={location.pathname} />

          <div className="ml-auto hidden sm:flex items-center gap-3">
            {tenant && (
              <span className="text-xs text-gray-500 bg-gray-100 px-2.5 py-1 rounded-full">
                {tenant.organization_name}
              </span>
            )}
          </div>
        </header>

        <main className="flex-1 overflow-y-auto">{children}</main>
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
          className="p-1 rounded hover:bg-gray-100 text-gray-500 hover:text-gray-700"
        >
          <ChevronLeft className="w-5 h-5" />
        </Link>
      )}
      <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
    </div>
  );
}
