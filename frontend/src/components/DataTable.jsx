import { useCallback, useEffect, useRef } from "react";
import { twMerge } from "tailwind-merge";

import { useDebouncedCallback } from "@app/hooks/useDebouncedCallback";

const SkeletonRow = ({ columns }) => (
  <tr className="border-b border-solid border-mineshaft-600 last:border-b-0">
    {Array.from({ length: columns }).map((_, i) => (
      <td key={i} className="px-5 py-3">
        <div className="h-4 w-full animate-pulse rounded bg-mineshaft-700" />
      </td>
    ))}
  </tr>
);

const EmptyState = ({ message }) => (
  <tr>
    <td colSpan={100} className="px-5 py-12 text-center">
      <p className="text-sm text-gray-400">{message}</p>
    </td>
  </tr>
);

const ErrorState = ({ message, onRetry }) => (
  <tr>
    <td colSpan={100} className="px-5 py-12 text-center">
      <p className="text-sm text-red-400">{message}</p>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="mt-2 rounded bg-mineshaft-600 px-3 py-1 text-xs text-gray-300 hover:bg-mineshaft-500"
        >
          Retry
        </button>
      )}
    </td>
  </tr>
);

export const DataTable = ({
  columns,
  data,
  keyExtractor,
  isLoading,
  isLoadingMore,
  hasMore,
  loadMore,
  error,
  onRetry,
  emptyMessage = "No data available.",
  errorMessage = "Failed to load data.",
  className,
  debounceDelay = 300
}) => {
  const sentinelRef = useRef(null);
  const isLoadingMoreRef = useRef(isLoadingMore);
  const hasMoreRef = useRef(hasMore);

  isLoadingMoreRef.current = isLoadingMore;
  hasMoreRef.current = hasMore;

  const debouncedLoadMore = useDebouncedCallback(
    useCallback(() => {
      if (!isLoadingMoreRef.current && hasMoreRef.current) {
        loadMore();
      }
    }, [loadMore]),
    debounceDelay
  );

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          debouncedLoadMore();
        }
      },
      { rootMargin: "200px" }
    );

    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, [debouncedLoadMore]);

  return (
    <div
      className={twMerge(
        "relative w-full overflow-x-auto rounded-lg border border-solid border-mineshaft-700 bg-mineshaft-800",
        className
      )}
    >
      <table className="w-full text-left text-sm text-gray-300">
        <thead className="bg-mineshaft-800 text-xs uppercase text-bunker-300">
          <tr>
            {columns.map((col) => (
              <th
                key={col.key}
                className={twMerge(
                  "border-b-2 border-mineshaft-600 bg-mineshaft-800 px-5 pt-4 pb-3.5 font-medium",
                  col.className
                )}
                style={col.width ? { width: col.width } : undefined}
              >
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {isLoading && data.length === 0 && (
            <>
              {Array.from({ length: 5 }).map((_, i) => (
                <SkeletonRow key={`skeleton-${i}`} columns={columns.length} />
              ))}
            </>
          )}

          {error && data.length === 0 && (
            <ErrorState message={errorMessage} onRetry={onRetry} />
          )}

          {!isLoading && !error && data.length === 0 && (
            <EmptyState message={emptyMessage} />
          )}

          {data.map((item) => (
            <tr
              key={keyExtractor(item)}
              className="border-b border-solid border-mineshaft-600 last:border-b-0 hover:bg-mineshaft-700"
            >
              {columns.map((col) => (
                <td
                  key={col.key}
                  className={twMerge("px-5 py-3", col.cellClassName)}
                >
                  {col.render ? col.render(item) : item[col.key]}
                </td>
              ))}
            </tr>
          ))}

          {isLoadingMore && (
            <>
              {Array.from({ length: 3 }).map((_, i) => (
                <SkeletonRow key={`more-skeleton-${i}`} columns={columns.length} />
              ))}
            </>
          )}

          {!hasMore && data.length > 0 && (
            <tr>
              <td colSpan={columns.length} className="px-5 py-4 text-center">
                <p className="text-xs text-gray-500">All data loaded.</p>
              </td>
            </tr>
          )}
        </tbody>
      </table>

      {hasMore && !isLoadingMore && (
        <div ref={sentinelRef} className="h-1 w-full" />
      )}
    </div>
  );
};