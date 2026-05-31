/**
 * UserListFilters — search input with debounce + status tab bar
 */

import { useState, useEffect, useCallback } from "react";
import { Input } from "../ui/input";
import type { UserListFilters as Filters, UserStatus } from "../../types/user";
import { Search } from "lucide-react";

interface UserListFiltersProps {
  filters: Filters;
  onFiltersChange: (filters: Filters) => void;
}

const statusTabs: { label: string; value: UserStatus | "" }[] = [
  { label: "All", value: "" },
  { label: "Active", value: "active" },
  { label: "Pending", value: "pending" },
  { label: "Disabled", value: "disabled" },
];

export function UserListFilters({
  filters,
  onFiltersChange,
}: UserListFiltersProps) {
  const [search, setSearch] = useState(filters.search ?? "");

  // Debounce search input 300ms
  useEffect(() => {
    const t = setTimeout(() => {
      onFiltersChange({ ...filters, search, page: 1 });
    }, 300);
    return () => clearTimeout(t);
    // Only re-run when the local search value changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [search]);

  const handleStatusChange = useCallback(
    (status: UserStatus | "") => {
      onFiltersChange({ ...filters, status, page: 1 });
    },
    [filters, onFiltersChange],
  );

  return (
    <div className="space-y-4">
      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
        <Input
          placeholder="Search by name or email..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Status tabs */}
      <div className="flex gap-1 border-b border-black">
        {statusTabs.map((tab) => (
          <button
            key={tab.value}
            onClick={() => handleStatusChange(tab.value)}
            className={`px-4 py-2 font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] border-b-2 transition-colors ${
              (filters.status ?? "") === tab.value
                ? "border-black text-black"
                : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>
    </div>
  );
}
