import { redirect } from '@sveltejs/kit';

import { authApi } from '$lib/api/auth.api';
import { ApiError } from '$lib/api/http-client';
import { type SessionUser } from '$lib/contracts/user.schema';
import { loginPathFor } from '$lib/navigation/login-return';
import { sidebarOpenFromCookie } from '$lib/utils/sidebar-state.util';

/**
 * A GUARDA de sessão: todo o resto do sistema mora dentro deste grupo de rota, e ninguém
 * chega em nenhuma tela dele sem sessão válida.
 *
 * `/api/auth/me` é a mesma origem de verdade que o backend usa para tudo o mais — não existe
 * um segundo lugar guardando "está logado" que pudesse divergir dele.
 *
 * Quem chega sem sessão vai para o LOGIN, levando o endereço em que estava (`?next=`): todo
 * depois de entrar a pessoa volta para a tela que pediu.
 *
 * Só 401 vira redirecionamento. Qualquer outro erro (rede caída, servidor fora) sobe para a
 * tela de erro padrão do SvelteKit — confundir "API fora do ar" com "não estou logado"
 * tiraria a pessoa da tela por um problema que login nenhum resolve.
 */
export async function load({
  url,
}: {
  url: URL;
}): Promise<{ user: SessionUser; sidebarOpen: boolean }> {
  try {
    /* A barra lateral lembra se estava aberta pelo cookie do `SidebarProvider` — lido no
       navegador, porque a tela é um SPA estático, sem servidor do SvelteKit. */
    return { user: await authApi.me(), sidebarOpen: sidebarOpenFromCookie(document.cookie) };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) redirect(307, loginPathFor(url));
    throw error;
  }
}
