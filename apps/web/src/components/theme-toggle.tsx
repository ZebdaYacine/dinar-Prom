'use client';

import {useEffect, useState} from 'react';
import {Icon} from './icon';

export function ThemeToggle() {
  const [dark, setDark] = useState(true);

  useEffect(() => {
    const saved = window.localStorage.getItem('dinar_theme');
    const next = saved !== 'light';
    setDark(next);
    document.documentElement.dataset.theme = next ? 'dark' : 'light';
  }, []);

  function toggle() {
    const next = !dark;
    setDark(next);
    window.localStorage.setItem('dinar_theme', next ? 'dark' : 'light');
    document.documentElement.dataset.theme = next ? 'dark' : 'light';
  }

  return <button className="icon-button" type="button" onClick={toggle} aria-label={dark ? 'Switch to light mode' : 'Switch to dark mode'} title={dark ? 'Light mode' : 'Dark mode'}><Icon name={dark ? 'sun' : 'moon'} /></button>;
}
