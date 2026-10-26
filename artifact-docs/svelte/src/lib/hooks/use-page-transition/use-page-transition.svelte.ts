import { afterNavigate, onNavigate } from '$app/navigation';

import { pageEnter, pageLeave } from '$lib/motion/motion';

/**
 * A TROCA DE PÁGINA: a tela atual sai e a nova entra em cascata.
 *
 * `onNavigate` segura a navegação até a saída terminar (ela é curta, ver `pageLeave`);
 * `afterNavigate` anima a entrada quando a tela nova já está no DOM — inclusive na primeira
 * abertura do sistema.
 *
 * Só anima quando a ROTA muda: trocar um filtro que mexe na query string, ou ir de
 * `/artifacts/a` para `/artifacts/a#versoes`, não é trocar de página, e a tela piscando ali
 * pareceria recarga.
 *
 * O elemento chega por função (`getTarget`) porque o `bind:this` do layout só é preenchido
 * depois que este hook já foi chamado.
 */
export function usePageTransition(getTarget: () => HTMLElement | null | undefined): void {
  onNavigate((navigation) => {
    if (navigation.from?.url.pathname === navigation.to?.url.pathname) return;

    return pageLeave(getTarget() ?? null);
  });

  afterNavigate(() => {
    pageEnter(getTarget() ?? null);
  });
}
