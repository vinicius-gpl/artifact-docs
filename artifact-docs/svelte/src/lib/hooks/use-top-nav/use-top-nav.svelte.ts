import { page } from '$app/state';

import { activeNavKeyOf, HOME_PATH, NAV_ITEMS, type NavItem } from '$lib/navigation/navigation';

export type TopNavModel = {
  data: { items: readonly NavItem[]; homeHref: string };
  state: { activeKey: string | undefined };
};

/**
 * O que a barra do topo precisa: os itens do menu e qual está aceso.
 *
 * Qual item acende é decidido AQUI, e não no componente: ler a rota atual é trabalho de hook,
 * e a barra continua sem saber de roteamento. `get` e não valor: o model é montado uma vez,
 * mas `page` muda a cada navegação.
 */
export function useTopNavModel(items: readonly NavItem[] = NAV_ITEMS): TopNavModel {
  return {
    data: { items, homeHref: HOME_PATH },
    get state() {
      return { activeKey: activeNavKeyOf(page.url.pathname, items) };
    },
  };
}
