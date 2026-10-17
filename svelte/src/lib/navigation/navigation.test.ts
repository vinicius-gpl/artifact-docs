import ListChecks from '@lucide/svelte/icons/list-checks';
import { describe, expect, it } from 'vitest';

import { activeNavKeyOf, HOME_PATH, NAV_ITEMS, type NavItem } from './navigation';

const items: readonly NavItem[] = [
  { key: 'artifacts', label: 'Artefatos', to: '/artifacts', icon: ListChecks },
  { key: 'artifacts/archive', label: 'Arquivo', to: '/artifacts/archive', icon: ListChecks },
];

describe('NAV_ITEMS', () => {
  // feliz
  it('gives every item a unique key that matches its path', () => {
    const keys = NAV_ITEMS.map((item) => item.key);

    expect(new Set(keys).size).toBe(keys.length);
    for (const item of NAV_ITEMS) expect(item.to).toBe(`/${item.key}`);
  });

  // feliz
  it('opens the system on the first item', () => {
    expect(HOME_PATH).toBe(NAV_ITEMS[0]?.to);
  });
});

describe('activeNavKeyOf', () => {
  // feliz
  it('lights the item on its own path and on its subroutes', () => {
    expect(activeNavKeyOf('/artifacts', items)).toBe('artifacts');
    expect(activeNavKeyOf('/artifacts/briefing', items)).toBe('artifacts');
  });

  // feliz
  it('prefers the longest matching prefix', () => {
    expect(activeNavKeyOf('/artifacts/archive/2', items)).toBe('artifacts/archive');
  });

  // triste
  it('lights nothing on a path outside the menu, or on a mere name prefix', () => {
    expect(activeNavKeyOf('/login', items)).toBeUndefined();
    expect(activeNavKeyOf('/artifactsx', items)).toBeUndefined();
  });
});
