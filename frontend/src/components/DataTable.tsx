import { useCallback, useEffect, useRef, useState } from "react";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@app/components/v3";
import { useDebouncedCallback } from "@app/hooks/useDebouncedCallback";

type Column<T> = {
  key: keyof T & string;
  header: string;
  render?: (row: T) => React.ReactNode;
  isTruncatable?: boolean;
};

type Props<T> = {
  columns: Column<T>[];
  fetchData: (page: number, pageSize: number) => Promise<{ data: T[]; hasMore: boolean }>;
  pageSize?: number;
  debounceDelay?: number;
  rowKey: keyof T & string;
  onRowClick?: (row: T) => void;
  emptyMessage?: string;
};

export const DataTable = <T extends Record<string, unknown>>({
  columns,
  fetchData,
  pageSize = 20,
  debounceDelay = 300,
  rowKey,
  onRowClick,
  emptyMessage = "No data available"
}: Props<T>) => {
  const [rows, setRows] = useState<T[]>([]);
  const [page, setPage] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const sentinelRef = useRef<HTMLTableRowElement>(null);
  const isInitialMount = useRef(true);

  const loadMore = useCallback(async () => {
    if (isLoading || isLoadingMore || !hasMore) return;

    const isInitial = isInitialMount.current;
    if (isInitial) {
      setIsLoading(true);
    } else {
      setIsLoadingMore(true);
    }

    try {
      const result = await fetchData(page, pageSize);
      setRows((prev) => (isInitial ? result.data : [...prev, ...result.data]));
      setHasMore(result.hasMore);
      setPage((p) => p + 1);
    } catch {
      // ignore
    } finally {
      setIsLoading(false);
      setIsLoadingMore(false);
      if (isInitialMount.current) {
        isInitialMount.current = false;
      }
    }
  }, [fetchData, page, pageSize, hasMore, isLoading, isLoadingMore]);

  const debouncedLoadMore = useDebouncedCallback(loadMore, debounceDelay);

  useEffect(() => {
    if (isInitialMount.current) {
      loadMore();
    }
  }, [loadMore]);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && hasMore && !isLoading && !isLoadingMore) {
          debouncedLoadMore();
        }
      },
      { rootMargin: "200px" }
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [debouncedLoadMore, hasMore, isLoading, isLoadingMore]);

  return (
    <Table>
      <TableHeader>
        <TableRow>
          {columns.map((col) => (
            <TableHead key={col.key} isTruncatable={col.isTruncatable}>
              {col.header}
            </TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.length === 0 && !isLoading && (
          <TableRow>
            <TableCell colSpan={columns.length} className="text-center text-muted">
              {emptyMessage}
            </TableCell>
          </TableRow>
        )}
        {rows.map((row) => (
          <TableRow
            key={String(row[rowKey])}
            onClick={onRowClick ? () => onRowClick(row) : undefined}
          >
            {columns.map((col) => (
              <TableCell key={col.key} isTruncatable={col.isTruncatable}>
                {col.render ? col.render(row) : String(row[col.key] ?? "")}
              </TableCell>
            ))}
          </TableRow>
        ))}
        {hasMore && (
          <TableRow ref={sentinelRef}>
            <TableCell colSpan={columns.length} className="text-center text-muted">
              {isLoadingMore ? "Loading more..." : ""}
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  );
};
