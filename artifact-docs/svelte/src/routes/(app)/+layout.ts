import { redirect } from '@sveltejs/kit';

import { authApi } from '$lib/api/auth.api';
import { ApiError } from '$lib/api/http-client';
import { type SessionUser } from '$lib/contracts/user.schema';
import { deniedLoginPath, loginPathFor } from '$lib/navigation/login-return';

/**
 * A GUARDA de sessão: as telas do hub moram dentro deste grupo de rota, e ninguém chega nelas
 * sem sessão válida.
 *
 * `/api/auth/me` é a mesma fonte de verdade que a API usa para tudo o mais (no modo mock,
 * quem responde é o `mock-server`, com o usuário de mentira).
 *
 * - 401 (sem sessão): login, levando o endereço em que estava (`?next=`).
 * - 403 (entrou no GitHub, mas a conta não está na lista): login, com o aviso de acesso negado.
 * - Qualquer outro erro sobe para a tela de erro: API fora do ar não é "não estou logado".
 */
export async function load({ url }: { url: URL }): Promise<{ user: SessionUser }> {
  try {
    return { user: await authApi.me() };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) redirect(307, loginPathFor(url));
    if (error instanceof ApiError && error.status === 403) redirect(307, deniedLoginPath());
    throw error;
  }
}
