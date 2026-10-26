import { redirect } from '@sveltejs/kit';

import { authApi } from '$lib/api/auth.api';
import {
  LOGIN_ERROR_PARAM,
  LOGIN_RETURN_PARAM,
  safeReturnPath,
} from '$lib/navigation/login-return';

/**
 * Quem já está logado não vê o login de novo — vai direto para onde ia (`?next=`).
 *
 * `.catch(() => null)`, e não `try/catch`: o `redirect()` do SvelteKit lança um objeto especial
 * que precisa ATRAVESSAR esta função, e um `try/catch` em volta dele o engoliria.
 */
export async function load({
  url,
}: {
  url: URL;
}): Promise<{ returnTo: string; error: string | null; needsSetup: boolean }> {
  const returnTo = safeReturnPath(url.searchParams.get(LOGIN_RETURN_PARAM));
  const error = url.searchParams.get(LOGIN_ERROR_PARAM);
  const needsSetup = await readNeedsSetup();
  /* Com o aviso de acesso negado, não confere a sessão: ela existe na Neon, mas a API a recusa —
     conferir de novo só devolveria o mesmo 403. */
  if (error) return { returnTo, error, needsSetup };

  const user = await authApi.me().catch(() => null);
  if (user) redirect(307, returnTo);

  return { returnTo, error: null, needsSetup };
}

/**
 * Ainda não existe a conta do dono? Na dúvida (API fora do ar), NÃO oferece o primeiro acesso:
 * um botão de cadastro aparecendo por engano é pior que ele sumir por um instante.
 */
async function readNeedsSetup(): Promise<boolean> {
  const status = await authApi.setup().catch(() => null);

  return status?.needsSetup ?? false;
}
