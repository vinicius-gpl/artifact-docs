import {
  type FirstAccessInput,
  type LoginInput,
  type SetupStatus,
} from '$lib/contracts/auth.schema';
import { type SessionUser } from '$lib/contracts/user.schema';

import { apiRequest } from './http-client';

/**
 * As perguntas de identidade à NOSSA API. Entrar e sair acontecem no Neon Auth
 * (`lib/auth/neon-auth.client.ts`); o que só a API sabe é se esta pessoa pode usar o hub.
 *
 * `mock*` só existem no modo mock (localhost): é o login com o usuário de mentira e o primeiro
 * acesso de mentira. Em produção essas rotas não existem no Go — lá, entrar e criar a conta
 * acontecem no Neon Auth.
 */
/**
 * As credenciais do usuário de teste do modo mock — valores de teste, nunca de uma conta real.
 * Moram aqui, e não nos fixtures, porque a tela de login as usa para vir preenchida: importar os
 * fixtures inteiros só por isso arrastaria os dados de mentira para o bundle.
 */
export const MOCK_CREDENTIALS = { email: 'voce@exemplo.com', password: 'artifact-docs' } as const;

export const authApi = {
  /** Quem está logado — ou lança 401 (sem sessão) / 403 (conta sem acesso). */
  me: () => apiRequest<SessionUser>('/auth/me'),
  /** Pública: ainda não existe a conta do dono? Só então a tela oferece o primeiro acesso. */
  setup: () => apiRequest<SetupStatus>('/auth/setup'),
  mockSignUp: (account: FirstAccessInput) =>
    apiRequest<SessionUser>('/auth/mock/sign-up', { method: 'POST', body: account }),
  mockSignIn: (credentials: LoginInput) =>
    apiRequest<SessionUser>('/auth/mock/sign-in', { method: 'POST', body: credentials }),
  mockSignOut: () => apiRequest<void>('/auth/mock/sign-out', { method: 'POST' }),
};
