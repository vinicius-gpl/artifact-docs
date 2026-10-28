import { type LoginInput, type SetupStatus } from '$lib/contracts/auth.schema';
import { type SessionUser } from '$lib/contracts/user.schema';

import { apiRequest } from './http-client';

/**
 * As perguntas de identidade à NOSSA API. Entrar e sair acontecem no Neon Auth
 * (`lib/auth/neon-auth.client.ts`); o que só a API sabe é se esta pessoa pode usar o hub.
 *
 * `mock*` só existem no modo mock (localhost): é o login com o usuário de mentira e o primeiro
 * acesso de mentira. Em produção essas rotas não existem no Go — lá, entrar e criar a senha
 * acontecem no Neon Auth.
 */
/**
 * As credenciais do usuário de teste do modo mock — valores de teste, nunca de uma conta real.
 * Moram aqui, e não nos fixtures, porque a tela de login as usa para vir preenchida: importar os
 * fixtures inteiros só por isso arrastaria os dados de mentira para o bundle.
 */
export const MOCK_CREDENTIALS = { email: 'voce@exemplo.com', password: 'artifact-docs' } as const;

/**
 * O "link do e-mail" do modo mock: não há e-mail em localhost, então a tela vai direto para a
 * senha nova com este token, que carrega o e-mail da conta.
 */
export const MOCK_LINK_PREFIX = 'mock:';

export const authApi = {
  /** Quem está logado — ou lança 401 (sem sessão) / 403 (conta sem acesso). */
  me: () => apiRequest<SessionUser>('/auth/me'),
  /** Pública: ainda não existe a conta do dono? */
  setup: () => apiRequest<SetupStatus>('/auth/setup'),
  mockSetPassword: (input: { token: string; password: string }) =>
    apiRequest<void>('/auth/mock/set-password', { method: 'POST', body: input }),
  mockSignIn: (credentials: LoginInput) =>
    apiRequest<SessionUser>('/auth/mock/sign-in', { method: 'POST', body: credentials }),
  mockSignOut: () => apiRequest<void>('/auth/mock/sign-out', { method: 'POST' }),
};
