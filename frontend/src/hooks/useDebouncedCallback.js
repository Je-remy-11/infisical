import { useRef, useEffect, useCallback } from 'react';

export function useDebouncedCallback(callback, delay) {
  const timeoutRef = useRef(null);
  const callbackRef = useRef(callback);

  // 保持 callback 的最新引用，避免闭包陷阱
  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  const debouncedCallback = useCallback((...args) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    
    timeoutRef.current = setTimeout(() => {
      if (callbackRef.current) {
        callbackRef.current(...args);
      }
    }, delay);
  }, [delay]);

  // 组件卸载时清理定时器
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return debouncedCallback;
}
