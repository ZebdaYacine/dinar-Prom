'use client';

import {useEffect, useId, useRef, useState} from 'react';

export type ComboOption = {value: string; label: string};
export function Combobox({label, placeholder, value, options, onChange, required = false, disabled = false}: {label: string; placeholder: string; value: string; options: ComboOption[]; onChange: (value: string) => void; required?: boolean; disabled?: boolean}) {
  const [open, setOpen] = useState(false); const [query, setQuery] = useState(''); const ref = useRef<HTMLDivElement>(null); const listId = useId();
  const selected = options.find((option) => option.value === value); const filtered = options.filter((option) => option.label.toLocaleLowerCase().includes(query.toLocaleLowerCase())).slice(0, 80);
  useEffect(() => { const close = (event: MouseEvent) => { if (ref.current && !ref.current.contains(event.target as Node)) setOpen(false); }; document.addEventListener('mousedown', close); return () => document.removeEventListener('mousedown', close); }, []);
  return <div className="combobox" ref={ref}><label>{label}<div className="combobox-input-wrap"><input role="combobox" aria-controls={listId} aria-expanded={disabled ? false : open} aria-autocomplete="list" required={required} disabled={disabled} value={open ? query : selected?.label ?? ''} placeholder={placeholder} onFocus={() => { if (!disabled) { setOpen(true); setQuery(''); } }} onChange={(event) => { setQuery(event.target.value); setOpen(true); }} onKeyDown={(event) => { if (event.key === 'Escape') setOpen(false); if (event.key === 'Enter' && filtered[0]) { event.preventDefault(); onChange(filtered[0].value); setOpen(false); } }} /><span aria-hidden="true">⌄</span></div></label>{open && !disabled && <div id={listId} className="combobox-list" role="listbox">{filtered.length ? filtered.map((option) => <button type="button" role="option" aria-selected={option.value === value} key={option.value} onClick={() => { onChange(option.value); setOpen(false); setQuery(''); }}>{option.label}</button>) : <span className="combobox-empty">—</span>}</div>}</div>;
}
