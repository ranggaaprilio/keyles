/**
 * Toast Notification Component
 */

import { useCallback, useEffect, useState } from 'react';
import { CheckCircle, XCircle, Info, X } from 'lucide-react';

export type ToastType = 'success' | 'error' | 'info';

interface ToastProps {
  id: string;
  type: ToastType;
  message: string;
  duration?: number;
  onClose: (id: string) => void;
}

export function Toast({ id, type, message, duration = 5000, onClose }: ToastProps) {
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsVisible(false);
      setTimeout(() => onClose(id), 300);
    }, duration);

    return () => clearTimeout(timer);
  }, [id, duration, onClose]);

  const icons = {
    success: <CheckCircle className="h-4 w-4 text-success" />,
    error: <XCircle className="h-4 w-4 text-destructive" />,
    info: <Info className="h-4 w-4 text-muted-foreground" />,
  };

  const borderColors = {
    success: 'border-l-4 border-l-success',
    error: 'border-l-4 border-l-destructive',
    info: 'border-l-4 border-l-muted-foreground',
  };

  return (
    <div
      className={`
        flex items-center gap-3 rounded-lg border border-border bg-background px-4 py-3
        shadow-elevated transition-all duration-slow
        ${borderColors[type]}
        ${isVisible ? 'animate-slide-in' : 'animate-slide-out'}
      `}
    >
      {icons[type]}
      <p className="flex-1 text-sm text-foreground">
        {message}
      </p>
      <button
        onClick={() => onClose(id)}
        className="text-muted-foreground hover:text-foreground"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}

interface ToastContainerProps {
  toasts: Array<{
    id: string;
    type: ToastType;
    message: string;
  }>;
  onClose: (id: string) => void;
}

export function ToastContainer({ toasts, onClose }: ToastContainerProps) {
  return (
    <div className="fixed top-4 right-4 z-50 space-y-2">
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          id={toast.id}
          type={toast.type}
          message={toast.message}
          onClose={onClose}
        />
      ))}
    </div>
  );
}

// Toast manager hook
export function useToast() {
  const [toasts, setToasts] = useState<Array<{
    id: string;
    type: ToastType;
    message: string;
  }>>([]);

  // Stable identities: callers put these in effect deps, so a new function
  // per render would re-run the effect (and add a toast) on every render.
  const addToast = useCallback((type: ToastType, message: string) => {
    const id = Math.random().toString(36).substring(2, 9);
    setToasts((prev) => [...prev, { id, type, message }]);
  }, []);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const success = useCallback((message: string) => addToast('success', message), [addToast]);
  const error = useCallback((message: string) => addToast('error', message), [addToast]);
  const info = useCallback((message: string) => addToast('info', message), [addToast]);

  return { toasts, addToast, removeToast, success, error, info };
}
