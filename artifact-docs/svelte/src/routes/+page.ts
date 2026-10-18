import { redirect } from '@sveltejs/kit';

import { HOME_PATH } from '$lib/navigation/navigation';

/**
 * A raiz não tem tela própria: ela manda para a primeira tela do menu (Artefatos).
 *
 * Quem não tem sessão é barrado pela guarda de `(app)/+layout.ts` ao chegar lá, e volta para
 * a mesma tela depois de entrar.
 */
export function load(): never {
  redirect(307, HOME_PATH);
}
