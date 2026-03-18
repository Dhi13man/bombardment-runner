import { createContext } from 'preact';
import { useState, useCallback, useContext, useRef } from 'preact/hooks';
import type { ComponentChildren } from 'preact';
import { Icon, type IconName } from '../Icon';

/* ---------- Types ---------- */

export type ToastType = 'success' | 'error' | 'warning' | 'info';

interface ToastItem {
  id: number;
  type: ToastType;
  message: string;
  exiting: boolean;
  duration: number;
}

interface ToastContextValue {
  showToast: (type: ToastType, message: string, duration?: number) => void;
}

/* ---------- Context ---------- */

const ToastContext = createContext<ToastContextValue>({
  showToast: () => {},
});

export function useToast(): ToastContextValue {
  return useContext(ToastContext);
}

/* ---------- Provider ---------- */

const TOAST_ICON: Record<ToastType, IconName> = {
  success: 'check-circle-2',
  error: 'alert-triangle',
  warning: 'alert-triangle',
  info: 'info',
};

export function ToastProvider({ children }: { children: ComponentChildren }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const nextId = useRef(0);

  const dismiss = useCallback((id: number) => {
    // Trigger exit animation
    setToasts((prev) => prev.map((t) => (t.id === id ? { ...t, exiting: true } : t)));
    // Remove after animation completes
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 200);
  }, []);

  const showToast = useCallback(
    (type: ToastType, message: string, duration = 5000) => {
      const id = nextId.current++;
      setToasts((prev) => [...prev, { id, type, message, exiting: false, duration }]);

      if (duration > 0) {
        setTimeout(() => dismiss(id), duration);
      }
    },
    [dismiss],
  );

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      {toasts.length > 0 && (
        <div class="toast-container" aria-live="polite">
          {toasts.map((toast) => (
            <div
              key={toast.id}
              class={`toast toast-${toast.type}${toast.exiting ? ' toast-exiting' : ''}`}
              role={toast.type === 'error' ? 'alert' : 'status'}
            >
              <Icon name={TOAST_ICON[toast.type]} class="toast-icon" />
              <span class="toast-message">{toast.message}</span>
              <button
                type="button"
                class="toast-close"
                aria-label="Dismiss"
                onClick={() => dismiss(toast.id)}
              >
                <Icon name="x" size="sm" />
              </button>
              <div
                class="toast-countdown"
                style={{ animationDuration: `${toast.duration}ms` }}
              />
            </div>
          ))}
        </div>
      )}
    </ToastContext.Provider>
  );
}
