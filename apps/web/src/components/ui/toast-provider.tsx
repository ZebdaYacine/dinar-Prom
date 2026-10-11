'use client';

import {createContext, useCallback, useContext, useMemo, useState} from 'react';

type ToastTone = 'success' | 'error' | 'info';
type ToastItem = {id: number; message: string; tone: ToastTone};
type ToastContextValue = {showToast: (message: string, tone?: ToastTone) => void};
const ToastContext = createContext<ToastContextValue | null>(null);

export function ToastProvider({children}: {children: React.ReactNode}) {
  const [items, setItems] = useState<ToastItem[]>([]);
  const showToast = useCallback((message: string, tone: ToastTone = 'info') => {
    const id = Date.now() + Math.random();
    setItems((current) => [...current, {id, message, tone}]);
    window.setTimeout(() => setItems((current) => current.filter((item) => item.id !== id)), 3600);
  }, []);
  const value = useMemo(() => ({showToast}), [showToast]);
  return <ToastContext.Provider value={value}>{children}<div className="toast-stack" aria-live="polite" aria-atomic="true">{items.map((item) => <div key={item.id} className={`toast toast-${item.tone}`} role="status">{item.message}</div>)}</div></ToastContext.Provider>;
}

export function useToast() { const context = useContext(ToastContext); if (!context) throw new Error('useToast must be used inside ToastProvider'); return context; }
