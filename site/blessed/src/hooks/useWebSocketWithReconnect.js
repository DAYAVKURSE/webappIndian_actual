import { useEffect, useRef } from 'react';

export const useWebSocketWithReconnect = (url, onMessage, reconnectDelay = 3000) => {
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const isUnmountedRef = useRef(false);

  useEffect(() => {
    const connect = () => {
      if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) return;
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected');
      };

      ws.onerror = (error) => {
        console.warn('WebSocket error:', error);
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected. Attempting reconnect...');
        if (!isUnmountedRef.current) {
          reconnectTimeoutRef.current = setTimeout(connect, reconnectDelay);
        }
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          onMessage?.(message);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };
    };

    connect();

    return () => {
      isUnmountedRef.current = true;
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [url, reconnectDelay, onMessage]);

  return wsRef;
};
