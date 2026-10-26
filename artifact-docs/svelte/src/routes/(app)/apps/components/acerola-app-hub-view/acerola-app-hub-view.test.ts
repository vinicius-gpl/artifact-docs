import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import { type AppGroup } from '$lib/hooks/use-app-hub/use-app-hub.svelte';
import AppHubView, { type AcerolaAppHubViewProps } from './acerola-app-hub-view.svelte';

const GROUP: AppGroup = {
  id: 't1',
  name: 'Self-hosted',
  apps: [
    {
      id: 'a1',
      name: 'Jellyfin',
      url: 'https://jellyfin.exemplo.com',
      host: 'jellyfin.exemplo.com',
      iconUrl: 'https://jellyfin.exemplo.com/favicon.ico',
      status: { label: 'No ar', tone: 'success' },
      checkedLabel: 'Nunca verificado',
    },
  ],
};

function props(
  data: Partial<AcerolaAppHubViewProps['data']> = {},
  state: Partial<AcerolaAppHubViewProps['state']> = {},
): AcerolaAppHubViewProps {
  return {
    data: { groups: [GROUP], total: 1, ...data },
    state: { isLoading: false, error: null, search: '', ...state },
    actions: { onRetry: vi.fn(), onClearSearch: vi.fn() },
  };
}

describe('AcerolaAppHubView', () => {
  // feliz
  it('shows each group with its apps', () => {
    render(AppHubView, { props: props() });

    expect(screen.getByRole('heading', { name: /Self-hosted/ })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Jellyfin/ })).toBeInTheDocument();
  });

  // triste
  it('shows the reason and retries when the hub fails', async () => {
    const view = props({}, { error: 'A API está fora do ar.' });
    render(AppHubView, { props: view });

    await fireEvent.click(screen.getByRole('button', { name: /Tentar de novo/ }));
    expect(view.actions.onRetry).toHaveBeenCalled();
  });

  it('says nothing is registered yet when there are no apps', () => {
    render(AppHubView, { props: props({ groups: [], total: 0 }) });

    expect(screen.getByText('Nenhum app cadastrado')).toBeInTheDocument();
  });

  it('offers to clear the search when it hid every app', async () => {
    const view = props({ groups: [], total: 3 }, { search: 'zzz' });
    render(AppHubView, { props: view });

    await fireEvent.click(screen.getByRole('button', { name: 'Limpar busca' }));
    expect(view.actions.onClearSearch).toHaveBeenCalled();
  });
});
