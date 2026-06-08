import { useCallback, useEffect, useRef } from "react";

import { useDebouncedCallback } from "../hooks/useDebouncedCallback";

const SCROLL_THRESHOLD = 120;

const resolveRowKey = (row, index) => row.id ?? row.key ?? index;

export default function DataTable({
  columns,
  rows,
  hasMore,
  isLoading,
  isLoadingMore = false,
  loadMore,
  emptyState = "No data available"
}) {
  const scrollContainerRef = useRef(null);
  const isLoadMoreInFlightRef = useRef(false);

  useEffect(() => {
    if (!isLoadingMore) {
      isLoadMoreInFlightRef.current = false;
    }
  }, [isLoadingMore]);

  const requestMoreRows = useCallback(async () => {
    if (!hasMore || isLoading || isLoadingMore || isLoadMoreInFlightRef.current) {
      return;
    }

    isLoadMoreInFlightRef.current = true;

    try {
      await Promise.resolve(loadMore());
    } finally {
      isLoadMoreInFlightRef.current = false;
    }
  }, [hasMore, isLoading, isLoadingMore, loadMore]);

  const debouncedLoadMore = useDebouncedCallback(requestMoreRows, 200, [requestMoreRows]);

  const handleScroll = useCallback(
    (event) => {
      const { scrollHeight, scrollTop, clientHeight } = event.currentTarget;
      const remainingScrollDistance = scrollHeight - scrollTop - clientHeight;

      if (remainingScrollDistance <= SCROLL_THRESHOLD) {
        debouncedLoadMore();
      }
    },
    [debouncedLoadMore]
  );

  useEffect(() => {
    const container = scrollContainerRef.current;

    if (!container) {
      return undefined;
    }

    const remainingScrollDistance = container.scrollHeight - container.scrollTop - container.clientHeight;

    if (remainingScrollDistance <= SCROLL_THRESHOLD) {
      debouncedLoadMore();
    }

    return undefined;
  }, [debouncedLoadMore, rows.length]);

  return (
    <div ref={scrollContainerRef} className="max-h-[600px] overflow-auto" onScroll={handleScroll}>
      <table className="min-w-full divide-y divide-mineshaft-700">
        <thead className="bg-mineshaft-800/70 sticky top-0 z-10">
          <tr>
            {columns.map((column) => (
              <th key={column.key} className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-bunker-300">
                {column.title}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-mineshaft-800 bg-mineshaft-900">
          {rows.length === 0 && !isLoading ? (
            <tr>
              <td className="px-4 py-6 text-center text-sm text-bunker-300" colSpan={columns.length}>
                {emptyState}
              </td>
            </tr>
          ) : (
            rows.map((row, index) => (
              <tr key={resolveRowKey(row, index)}>
                {columns.map((column) => (
                  <td key={column.key} className="whitespace-nowrap px-4 py-3 text-sm text-bunker-200">
                    {column.render ? column.render(row, index) : row[column.key]}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>

      {(isLoading || isLoadingMore) && (
        <div className="px-4 py-3 text-center text-sm text-bunker-300">Loading...</div>
      )}
    </div>
  );
}
