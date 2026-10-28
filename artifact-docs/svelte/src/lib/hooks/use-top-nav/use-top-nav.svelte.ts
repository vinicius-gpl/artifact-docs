import { goto } from '$app/navigation';
import { page } from '$app/state';
import { useQueryClient } from '@tanstack/svelte-query';

import { authApi } from '$lib/api/auth.api';
import { USE_MOCKS } from '$lib/api/http-client';
import { signOutOfNeon } from '$lib/auth/neon-auth.client';
import { markedTab } from '$lib/navigation/active-tab.svelte';
import { activeNavKeyOf, HOME_PATH, NAV_ITEMS, type NavItem } from '$lib/navigation/navigation';

export type TopNavModel = {
  data: {
    items: readonly NavItem[];
    homeHref: string;
  };
  state: { activeKey: string | undefined; isSigningOut: boolean };
  actions: { onSignOut: () => void };
};

/**
 * O que a barra do topo precisa: os itens, qual está aceso e como sair.
 *
 * Qual item acende é decidido AQUI, e não no componente: ler a rota atual é trabalho de hook.
 * `get` e não valor: o model é montado uma vez, mas `page` muda a cada navegação.
 *
 * Sair limpa o cache de dados (`queryClient.clear()`): quem abrir o hub depois nesta máquina
 * não vê, nem por um instante, a lista que estava na tela. No modo mock, sair fecha a sessão
 * de mentira; em produção, a sessão da Neon.
 */
export function useTopNavModel(input: { items?: readonly NavItem[] } = {}): TopNavModel {
  const queryClient = useQueryClient();
  const items = input.items ?? NAV_ITEMS;
  let isSigningOut = $state(false);

  async function signOut() {
    if (USE_MOCKS) await authApi.mockSignOut();
    else await signOutOfNeon();

    queryClient.clear();
    await goto('/login');
  }

  return {
    data: {
      items,
      homeHref: HOME_PATH,
    },
    get state() {
      /* A tela aberta pode dizer a aba (o detalhe de uma ideia acende Ideias, e não a aba do
         endereço); sem marca, vale o endereço. */
      const activeKey = markedTab() ?? activeNavKeyOf(page.url.pathname, items);

      return { activeKey, isSigningOut };
    },
    actions: {
      onSignOut: () => {
        isSigningOut = true;
        signOut().finally(() => {
          isSigningOut = false;
        });
      },
    },
  };
}
