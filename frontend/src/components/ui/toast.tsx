/**
 * Toast Notification Component
 */

import { useEffect, useState } from 'react';
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
    success: <CheckCircle className="h-4 w-4 text-green-700" />,
    error: <XCircle className="h-4 w-4 text-red-700" />,
    info: <Info className="h-4 w-4 text-blue-700" />,
  };

  const borderColors = {
    success: 'border-l-4 border-l-green-700',
    error: 'border-l-4 border-l-red-700',
    info: 'border-l-4 border-l-blue-700',
  };

  return (
    <div
      className={`
        flex items-center gap-3 border border-black bg-white px-4 py-3
        shadow-[2px_2px_0_#000] transition-all duration-300
        ${borderColors[type]}
        ${isVisible ? 'animate-slide-in' : 'animate-slide-out'}
      `}
    >
      {icons[type]}
      <p className="flex-1 font-['Times_New_Roman',Times,serif] text-sm text-black">
        {message}
      </p>
      <button
        onClick={() => onClose(id)}
        className="text-gray-500 hover:text-black"
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

  const addToast = (type: ToastType, message: string) => {
    const id = Math.random().toString(36).substring(2, 9);
    setToasts((prev) => [...prev, { id, type, message }]);
  };

  const removeToast = (id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  return {
    toasts,
    addToast,
    removeToast,
    success: (message: string) => addToast('success', message),
    error: (message: string) => addToast('error', message),
    info: (message: string) => addToast('info', message),
  };
}
