import { useEffect, useRef, useState } from 'react';

interface SSEOptions {
  url: string;
  enabled?: boolean;
  onMessage?: (data: any) => void;
  onError?: (error: Event) => void;
}

/**
 * Custom hook for Server-Sent Events (SSE) real-time updates
 */
export const useSSE = ({ url, enabled = true, onMessage, onError }: SSEOptions) => {
  const [isConnected, setIsConnected] = useState(false);
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    const token = localStorage.getItem('token');
    if (!token) {
      return;
    }

    // EventSource doesn't support custom headers, so pass token as query param
    const urlWithToken = `${url}${url.includes('?') ? '&' : '?'}token=${token}`;
    const eventSource = new EventSource(urlWithToken);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      setIsConnected(true);
    };

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage?.(data);
      } catch (error) {
        console.error('Failed to parse SSE message:', error);
      }
    };

    eventSource.onerror = (error) => {
      setIsConnected(false);
      onError?.(error);
      eventSource.close();
    };

    return () => {
      eventSource.close();
      setIsConnected(false);
    };
  }, [url, enabled, onMessage, onError]);

  return { isConnected };
};
