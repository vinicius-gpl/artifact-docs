import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import { type AppTile } from '$lib/hooks/use-app-hub/use-app-hub.svelte';
import AppGroup from './acerola-app-group.svelte';

function tile(id: string, name: string): AppTile {
  return {
    id,
    name,
    url: `https://${id}.exemplo.com`,
    host: `${id}.exemplo.com`,
    iconUrl: `https://${id}.exemplo.com/favicon.ico`,
    status: { label: 'No ar', tone: 'success' },
    checkedLabel: 'Nunca verificado',
  };
}

describe('AcerolaAppGroup', () => {
  // feliz
  it('names the group and lists its apps', () => {
    render(AppGroup, {
      props: {
        data: {
          group: {
            id: 't1',
            name: 'Self-hosted',
            apps: [tile('a', 'Jellyfin'), tile('b', 'Nextcloud')],
          },
        },
      },
    });

    expect(screen.getByRole('region', { name: /Self-hosted/ })).toBeInTheDocument();
    expect(screen.getAllByRole('link')).toHaveLength(2);
  });

  // triste
  it('still names a group with a single app (edge case)', () => {
    render(AppGroup, {
      props: { data: { group: { id: 't2', name: 'Na VPS', apps: [tile('c', 'Kuma')] } } },
    });

    expect(screen.getByRole('heading', { name: /Na VPS/ })).toHaveTextContent('1');
  });
});
