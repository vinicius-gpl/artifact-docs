import { createQuery } from '@tanstack/svelte-query';

import { appsApi } from '$lib/api/apps.api';
import { readError } from '$lib/api/http-client';
import { type StatusBadgeTone } from '$lib/components/acerola-status-badge/acerola-status-badge.svelte';
import {
  APP_LINK_STATUS_LABELS,
  type AppHub,
  type AppLink,
  type AppLinkStatus,
} from '$lib/contracts/app-link.schema';
import { mirrorStore } from '$lib/hooks/use-mirror-store/use-mirror-store.svelte';
import { formatTimeAgo } from '$lib/utils/format-machine';
import { siteOf } from '$lib/utils/site-url.util';

export type AppTile = {
  id: string;
  name: string;
  url: string;
  host: string;
  /** A logo enviada ou, sem ela, o favicon do próprio endereço. */
  iconUrl: string;
  status: { label: string; tone: StatusBadgeTone };
  checkedLabel: string;
};

export type AppGroup = { id: string; name: string; apps: AppTile[] };

export type AppHubModel = {
  data: { groups: AppGroup[]; total: number; offline: number };
  state: { isLoading: boolean; error: string | null; search: string };
  actions: { onSearch: (value: string) => void; onRetry: () => void };
};

export const APP_STATUS_TONES: Record<AppLinkStatus, StatusBadgeTone> = {
  online: 'success',
  offline: 'danger',
  unknown: 'neutral',
};

export function toAppTile(app: AppLink, now: number): AppTile {
  const { host, origin } = siteOf(app.url);

  return {
    id: app.id,
    name: app.name,
    url: app.url,
    host,
    iconUrl: app.logoUrl ?? (origin ? `${origin}/favicon.ico` : ''),
    status: { label: APP_LINK_STATUS_LABELS[app.status], tone: APP_STATUS_TONES[app.status] },
    checkedLabel: app.checkedAt
      ? `Verificado ${formatTimeAgo(app.checkedAt, now)}`
      : 'Nunca verificado',
  };
}

function matches(app: AppLink, search: string): boolean {
  const term = search.trim().toLocaleLowerCase('pt-BR');
  if (!term) return true;

  return `${app.name} ${app.url}`.toLocaleLowerCase('pt-BR').includes(term);
}

/**
 * Os apps agrupados pelo tipo, na ordem que a pessoa deu aos tipos (`position`). Tipo sem
 * nenhum app que passe na busca some — um título sem nada embaixo parece erro de carga.
 */
export function groupApps(hub: AppHub, search: string, now: number): AppGroup[] {
  return [...hub.types]
    .sort((a, b) => a.position - b.position)
    .map((type) => ({
      id: type.id,
      name: type.name,
      apps: hub.apps
        .filter((app) => app.typeId === type.id && matches(app, search))
        .map((app) => toAppTile(app, now)),
    }))
    .filter((group) => group.apps.length > 0);
}

/** O hub de apps: atalhos agrupados por tipo, com a situação que a goroutine do Go mediu. */
export function useAppHubModel(): AppHubModel {
  const hub = mirrorStore(createQuery({ queryKey: ['apps'], queryFn: () => appsApi.hub() }));
  let search = $state('');

  return {
    get data() {
      const value = hub.current.data ?? { types: [], apps: [] };

      return {
        groups: groupApps(value, search, Date.now()),
        total: value.apps.length,
        offline: value.apps.filter((app) => app.status === 'offline').length,
      };
    },
    get state() {
      return {
        isLoading: hub.current.isPending,
        error: readError(hub.current.error),
        search,
      };
    },
    actions: {
      onSearch: (value) => {
        search = value;
      },
      onRetry: () => {
        void hub.current.refetch();
      },
    },
  };
}
