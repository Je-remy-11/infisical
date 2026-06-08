import { useCallback, useRef } from "react";

export const useDebouncedCallback = <T extends (...args: unknown[]) => void>(
  callback: T,
  delay: number
): T => {
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const debouncedCallback = useCallback(
    (...args: Parameters<T>) => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
      timerRef.current = setTimeout(() => {
        callback(...args);
      }, delay);
    },
    [callback, delay]
  ) as T;

  return debouncedCallback;
};
