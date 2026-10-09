'use client';

import {useEffect, useState} from 'react';

function remaining(target: string) {
  return Math.max(0, new Date(target).getTime() - Date.now());
}

export function Countdown({endsAt, locale}: {endsAt: string; locale: string}) {
  const [milliseconds, setMilliseconds] = useState(() => remaining(endsAt));
  useEffect(() => {
    const timer = window.setInterval(() => setMilliseconds(remaining(endsAt)), 60_000);
    return () => window.clearInterval(timer);
  }, [endsAt]);
  const hours = Math.floor(milliseconds / 3_600_000);
  const days = Math.floor(hours / 24);
  const value = days > 0 ? `${days}d ${hours % 24}h` : `${hours}h`;
  return <span className="countdown" aria-label={locale === 'ar' ? `ينتهي خلال ${value}` : `Se termine dans ${value}`}>{locale === 'ar' ? `ينتهي خلال ${value}` : `Se termine dans ${value}`}</span>;
}
