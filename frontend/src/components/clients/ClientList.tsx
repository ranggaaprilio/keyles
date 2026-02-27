import { useState, useEffect, useCallback } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Search, ChevronLeft, ChevronRight, Plus } from 'lucide-react';
import { useClients } from '@/hooks/useClients';
import { ClientCard } from './ClientCard';

interface ClientListProps {
  onSelectClient: (clientId: string) => void;
  onCreateNew: () => void;
}

export function ClientList({ onSelectClient, onCreateNew }: ClientListProps) {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  const { data, isLoading, isError } = useClients({
    page,
    page_size: 10,
    search: debouncedSearch || undefined,
  });

  const handlePrev = useCallback(() => setPage((p) => Math.max(1, p - 1)), []);
  const handleNext = useCallback(() => {
    if (data && page < data.total_pages) setPage((p) => p + 1);
  }, [data, page]);

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Client Applications</h1>
        <Button onClick={onCreateNew}>
          <Plus className="h-4 w-4 mr-1" /> Register Client
        </Button>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search clients by name..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-24 w-full rounded-lg" />
          ))}
        </div>
      )}

      {/* Error */}
      {isError && (
        <div className="text-center py-8 text-destructive">
          Failed to load clients. Please try again.
        </div>
      )}

      {/* Empty state */}
      {!isLoading && !isError && data && data.clients.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          {debouncedSearch
            ? `No clients found matching "${debouncedSearch}"`
            : 'No client applications registered yet.'}
        </div>
      )}

      {/* Client list */}
      {!isLoading && data && data.clients.length > 0 && (
        <>
          <div className="space-y-3">
            {data.clients.map((client) => (
              <ClientCard
                key={client.client_id}
                client={client}
                onClick={onSelectClient}
              />
            ))}
          </div>

          {/* Pagination */}
          {data.total_pages > 1 && (
            <div className="flex items-center justify-between pt-2">
              <p className="text-sm text-muted-foreground">
                Page {data.page} of {data.total_pages} ({data.total} total)
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handlePrev}
                  disabled={page <= 1}
                >
                  <ChevronLeft className="h-4 w-4 mr-1" /> Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleNext}
                  disabled={page >= data.total_pages}
                >
                  Next <ChevronRight className="h-4 w-4 ml-1" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
