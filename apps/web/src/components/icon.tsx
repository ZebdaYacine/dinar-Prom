export type IconName = 'search' | 'filter' | 'bell' | 'sun' | 'moon' | 'house' | 'grid' | 'map-pin' | 'user' | 'languages' | 'menu' | 'x' | 'share' | 'whatsapp' | 'phone' | 'check' | 'clock' | 'arrow' | 'building' | 'car' | 'tire' | 'mobile' | 'laptop' | 'shirt' | 'sparkles' | 'sofa' | 'briefcase' | 'wrench' | 'utensils' | 'plug' | 'plane' | 'dumbbell' | 'flag-dz' | 'flag-fr';

const paths: Record<Exclude<IconName, 'flag-dz' | 'flag-fr'>, React.ReactNode> = {
  search: <><circle cx="11" cy="11" r="6.5"/><path d="m16 16 5 5"/></>,
  filter: <><path d="M4 6h16M7 12h10M10 18h4"/></>,
  bell: <><path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9"/><path d="M10 21h4"/></>,
  sun: <><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.42 1.42M17.65 17.65l1.42 1.42M2 12h2M20 12h2M4.93 19.07l1.42-1.42M17.65 6.35l1.42-1.42"/></>,
  moon: <path d="M20.5 15.2A8.5 8.5 0 0 1 8.8 3.5 8.5 8.5 0 1 0 20.5 15.2Z"/>,
  house: <><path d="m3 10 9-7 9 7v10H3Z"/><path d="M9 20v-6h6v6"/></>,
  grid: <><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></>,
  building: <><path d="M4 21V5l8-3 8 3v16"/><path d="M8 9h1M15 9h1M8 13h1M15 13h1M8 17h1M15 17h1M2 21h20"/></>,
  car: <><path d="m5 16 1.5-6h11L19 16"/><path d="M3 16h18v4H3zM7 20v1M17 20v1"/><circle cx="7" cy="16" r="1"/><circle cx="17" cy="16" r="1"/></>,
  tire: <><circle cx="12" cy="12" r="8"/><circle cx="12" cy="12" r="3"/><path d="M12 4v5M12 15v5M4 12h5M15 12h5"/></>,
  mobile: <><rect x="7" y="2.5" width="10" height="19" rx="2"/><path d="M10 5h4M11 18.5h2"/></>,
  laptop: <><path d="M5 5h14v10H5zM2 19h20M8 19l1-2h6l1 2"/></>,
  shirt: <><path d="m8 4 4 3 4-3 5 4-3 4-2-2v11H8V10l-2 2-3-4z"/></>,
  sparkles: <><path d="m12 3 1.4 5.1L18 10l-4.6 1.9L12 17l-1.4-5.1L6 10l4.6-1.9z"/><path d="m19 15 .7 2.3L22 18l-2.3.7L19 21l-.7-2.3L16 18l2.3-.7z"/></>,
  sofa: <><path d="M5 12V9a3 3 0 0 1 3-3h8a3 3 0 0 1 3 3v3"/><path d="M4 11a2 2 0 0 0-2 2v4h20v-4a2 2 0 0 0-2-2"/><path d="M4 17v3M20 17v3M6 15h12"/></>,
  briefcase: <><rect x="3" y="7" width="18" height="13" rx="2"/><path d="M8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M3 12h18M10 12v2h4v-2"/></>,
  wrench: <><path d="M14 6a5 5 0 0 0-6.4 6.4L3 17a2 2 0 1 0 2.8 2.8l4.6-4.6A5 5 0 0 0 16.8 9l-3.3 1.1-2.6-2.6z"/></>,
  utensils: <><path d="M7 3v8M4 3v5a3 3 0 0 0 6 0V3M7 11v10M16 3v18M16 3c3 2 3 6 0 8"/></>,
  plug: <><path d="M8 3v6M16 3v6M6 9h12v2a6 6 0 0 1-12 0V9ZM12 17v4M9 21h6"/></>,
  plane: <><path d="m3 11 18-6-6 18-3-8zM12 15 3 12l6-3"/></>,
  dumbbell: <><path d="M4 9v6M7 7v10M17 7v10M20 9v6M7 12h10M2 10v4M22 10v4"/></>,
  'map-pin': <><path d="M20 10c0 5-8 11-8 11S4 15 4 10a8 8 0 1 1 16 0Z"/><circle cx="12" cy="10" r="2.5"/></>,
  user: <><circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/></>,
  languages: <><path d="M4 5h10M9 3v2c0 5-2 8-5 10M6 10c2 2 4 3 7 4M15 12h6M18 9l-4 10M16 16h5"/></>,
  menu: <><path d="M4 6h16M4 12h16M4 18h16"/></>,
  x: <><path d="m5 5 14 14M19 5 5 19"/></>,
  share: <><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="m8.6 13.5 6.8 4M15.4 6.5 8.6 10.5"/></>,
  whatsapp: <><path d="M20 11.5a8 8 0 0 1-11.7 7L4 20l1.5-4.1A8 8 0 1 1 20 11.5Z"/><path d="M9 8.5c.3 2.2 1.5 3.8 3.8 4.7l1.4-1.3c.2-.2.5-.2.8-.1l1.5.7c.3.1.4.4.3.7-.3 1.2-1.2 1.7-2.3 1.6-3.7-.4-6.1-2.6-6.8-6.1-.2-1.1.4-2 1.5-2.2.3 0 .6.1.7.4l.7 1.5c.1.3.1.6-.1.8L9 8.5Z"/></>,
  phone: <path d="M7 4h3l1.5 4-2 1.5a14 14 0 0 0 3 3l1.5-2 4 1.5v3c0 1-1 2-2 1.9C9.5 16.2 5.8 12.5 5.1 6 5 5 6 4 7 4Z"/>,
  check: <><circle cx="12" cy="12" r="9"/><path d="m8 12 2.5 2.5L16 9"/></>,
  clock: <><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></>,
  arrow: <path d="M5 12h14M13 6l6 6-6 6"/>
};

export function Icon({name, size = 20, strokeWidth = 1.8, label}: {name: IconName; size?: number; strokeWidth?: number; label?: string}) {
  if (name === 'flag-dz') return <span className="flag flag-dz" aria-hidden="true"><i /></span>;
  if (name === 'flag-fr') return <span className="flag flag-fr" aria-hidden="true"><i /></span>;
  return <svg className="icon" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={strokeWidth} strokeLinecap="round" strokeLinejoin="round" role={label ? 'img' : undefined} aria-label={label}>{paths[name]}</svg>;
}
