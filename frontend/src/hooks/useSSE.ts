import { useEffect, useRef, useState } from 'react'

interface SSEOptions {
  onOpen?: () => void
  onMessage?: (event: MessageEvent) => void
  onError?: (error: Event) => void
  onClose?: () => void
}

interface UseSSEReturn {
  lastMessage: MessageEvent | null
  readyState: number
  error: Event | null
}

export const useSSE = (url: string, options: SSEOptions = {}): UseSSEReturn => {
  const [lastMessage, setLastMessage] = useState<MessageEvent | null>(null)
  const [readyState, setReadyState] = useState<number>(EventSource.CONNECTING)
  const [error, setError] = useState<Event | null>(null)
  const eventSourceRef = useRef<EventSource | null>(null)

  useEffect(() => {
    // Skip if no token available
    const token = localStorage.getItem('token')
    if (!token) return

    // Close existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    // Create new EventSource connection
    const baseUrl = 'http://localhost:8082'
    const fullUrl = `${baseUrl}${url}`
    const eventSource = new EventSource(fullUrl, {
      withCredentials: true,
    })

    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setReadyState(EventSource.OPEN)
      setError(null)
      options.onOpen?.()
    }

    eventSource.onmessage = (event) => {
      setLastMessage(event)
      options.onMessage?.(event)
    }

    eventSource.onerror = (event) => {
      setReadyState(EventSource.CLOSED)
      setError(event)
      options.onError?.(event)

      // Auto-reconnect after 5 seconds
      setTimeout(() => {
        if (eventSource.readyState === EventSource.CLOSED) {
          // Reconnection will be handled by the useEffect
        }
      }, 5000)
    }

    return () => {
      eventSource.close()
      eventSourceRef.current = null
      options.onClose?.()
    }
  }, [url, options])

  return { lastMessage, readyState, error }
}