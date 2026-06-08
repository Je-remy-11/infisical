import React, { useState, useEffect, useRef } from "react";
import { useDebouncedCallback } from "@app/hooks";

// 模拟数据获取函数
const fetchMoreData = async (page: number) => {
  console.log(`Fetching page ${page}...`);
  // 模拟网络延迟
  await new Promise((resolve) => setTimeout(resolve, 1000));
  // 返回新数据
  const newData = Array.from({ length: 10 }, (_, i) => ({
    id: (page - 1) * 10 + i + 1,
    text: `Item ${(page - 1) * 10 + i + 1}`
  }));
  return newData;
};

export const InfiniteScrollExample = () => {
  const [data, setData] = useState<Array<{ id: number; text: string }>>([]);
  const [page, setPage] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const lastItemRef = useRef<HTMLDivElement | null>(null);

  // 初始加载
  useEffect(() => {
    const loadInitialData = async () => {
      setIsLoading(true);
      const initialData = await fetchMoreData(1);
      setData(initialData);
      setPage(2);
      setIsLoading(false);
    };
    loadInitialData();
  }, []);

  // 使用 useDebouncedCallback 来防抖加载更多函数，延迟 300ms
  const debouncedLoadMore = useDebouncedCallback(async () => {
    if (isLoading || !hasMore) return;

    setIsLoading(true);
    try {
      const newData = await fetchMoreData(page);
      if (newData.length === 0) {
        setHasMore(false);
      } else {
        setData((prev) => [...prev, ...newData]);
        setPage((prev) => prev + 1);
      }
    } catch (error) {
      console.error("Failed to load more data:", error);
    } finally {
      setIsLoading(false);
    }
  }, 300);

  // 设置 IntersectionObserver 来检测何时加载更多
  useEffect(() => {
    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !isLoading) {
          debouncedLoadMore();
        }
      },
      { threshold: 0.1 }
    );

    const currentObserver = observerRef.current;
    const currentLastItem = lastItemRef.current;

    if (currentLastItem) {
      currentObserver.observe(currentLastItem);
    }

    return () => {
      if (currentLastItem) {
        currentObserver.unobserve(currentLastItem);
      }
    };
  }, [hasMore, isLoading, debouncedLoadMore]);

  return (
    <div className="p-4">
      <h1 className="text-xl font-bold mb-4">Infinite Scroll with useDebouncedCallback</h1>
      <div className="space-y-2">
        {data.map((item, index) => (
          <div
            key={item.id}
            ref={index === data.length - 1 ? lastItemRef : null}
            className="p-3 border rounded bg-white"
          >
            {item.text}
          </div>
        ))}
      </div>
      {isLoading && (
        <div className="p-4 text-center text-gray-500">Loading more...</div>
      )}
      {!hasMore && (
        <div className="p-4 text-center text-gray-500">No more items to load</div>
      )}
      <div className="mt-4 p-4 bg-gray-50 rounded">
        <h2 className="font-semibold mb-2">How it works:</h2>
        <ul className="list-disc list-inside space-y-1 text-sm">
          <li>useDebouncedCallback ensures loadMore is not called multiple times in quick succession</li>
          <li>It delays the function call by 300ms and cancels previous pending calls</li>
          <li>IntersectionObserver detects when the last item is visible and triggers the load</li>
        </ul>
      </div>
    </div>
  );
};
