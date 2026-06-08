import React, { useState, useEffect, useRef } from 'react';
import { useDebouncedCallback } from '../hooks/useDebouncedCallback';

const DataTable = ({ fetchMoreData, initialData = [] }) => {
  const [data, setData] = useState(initialData);
  const [loading, setLoading] = useState(false);
  const containerRef = useRef(null);

  // 原始的 loadMore 函数
  const loadMore = async () => {
    if (loading) return;
    setLoading(true);
    try {
      const newData = await fetchMoreData();
      if (newData && newData.length > 0) {
        setData((prev) => [...prev, ...newData]);
      }
    } catch (error) {
      console.error("Failed to load more data", error);
    } finally {
      setLoading(false);
    }
  };

  // 使用防抖 Hook 包装 loadMore，延迟 300ms
  const debouncedLoadMore = useDebouncedCallback(loadMore, 300);

  // 监听滚动事件
  const handleScroll = () => {
    if (!containerRef.current) return;
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
    
    // 触底检测 (距离底部 50px)
    if (scrollHeight - scrollTop - clientHeight < 50) {
      debouncedLoadMore();
    }
  };

  useEffect(() => {
    const container = containerRef.current;
    if (container) {
      container.addEventListener('scroll', handleScroll);
      return () => container.removeEventListener('scroll', handleScroll);
    }
  }, [debouncedLoadMore]);

  return (
    <div 
      ref={containerRef} 
      style={{ height: '400px', overflowY: 'auto', border: '1px solid #ccc' }}
    >
      <table style={{ width: '100%', textAlign: 'left' }}>
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Value</th>
          </tr>
        </thead>
        <tbody>
          {data.map((item, index) => (
            <tr key={`${item.id}-${index}`}>
              <td>{item.id}</td>
              <td>{item.name}</td>
              <td>{item.value}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {loading && <div style={{ padding: '10px', textAlign: 'center' }}>Loading...</div>}
    </div>
  );
};

export default DataTable;
