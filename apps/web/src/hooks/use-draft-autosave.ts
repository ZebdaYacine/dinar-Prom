'use client';

import {useEffect, useRef, useState} from 'react';

export type AutosaveState = 'idle' | 'saving' | 'saved' | 'error';

export function useDraftAutosave<T>(value: T, enabled: boolean, save: (value: T) => Promise<boolean>, delay = 10_000) {
  const saveRef = useRef(save);
  const lastSaved = useRef('');
  const [state, setState] = useState<AutosaveState>('idle');

  useEffect(() => { saveRef.current = save; }, [save]);
  useEffect(() => {
    if (!enabled) return;
    const snapshot = JSON.stringify(value);
    if (snapshot === lastSaved.current) return;
    const timer = window.setTimeout(() => {
      setState('saving');
      void saveRef.current(value).then((ok) => {
        if (ok) {
          lastSaved.current = snapshot;
          setState('saved');
        } else {
          setState('error');
        }
      }).catch(() => setState('error'));
    }, delay);
    return () => window.clearTimeout(timer);
  }, [delay, enabled, value]);

  return state;
}
