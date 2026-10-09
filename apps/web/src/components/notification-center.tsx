'use client';

import {Icon} from './icon';

export function NotificationCenter() {
  if (process.env.NEXT_PUBLIC_NOTIFICATIONS_ENABLED !== 'true') return null;
  return <button className="icon-button" type="button" aria-label="Notifications" title="Notifications">
    <Icon name="bell" />
    <span className="notification-dot" aria-label="New notifications" />
  </button>;
}
