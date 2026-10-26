import { redirect } from '@sveltejs/kit';

import { HOME_PATH } from '$lib/navigation/navigation';

/** A raiz não tem tela própria: ela manda para a primeira tela do menu (Artefatos). */
export function load(): never {
  redirect(307, HOME_PATH);
}
