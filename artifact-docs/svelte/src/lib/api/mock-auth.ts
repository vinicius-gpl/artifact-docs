import {
  INVALID_CREDENTIALS,
  loginRequestSchema,
  newPasswordSchema,
  setupStatusSchema,
  type SetupStatus,
} from '$lib/contracts/auth.schema';
import { sessionUserSchema, type SessionUser } from '$lib/contracts/user.schema';

import { MOCK_LINK_PREFIX } from './auth.api';
import { ApiError } from './http-client';
import { MOCK_PASSWORD, MOCK_USER } from './mock-fixtures';

/**
 * A IDENTIDADE DE MENTIRA do modo mock — o que o Neon Auth e o Go fazem em produção, imitado
 * no navegador para a tela de login funcionar sem conta nenhuma.
 *
 * Duas portas de entrada:
 * - o USUÁRIO DE TESTE (`MOCK_USER`), que existe sempre: em desenvolvimento ninguém precisa
 *   criar conta para abrir o hub;
 * - a CONTA DO PRIMEIRO ACESSO, que ganha senha pela tela de senha nova. Em produção o link
 *   vem por e-mail; aqui o token do link é `mock:<e-mail>` — é assim que esse caminho é
 *   testado em localhost.
 *
 * Sessão e conta ficam no `localStorage`: sem isso, cada F5 mandaria para o login. É valor de
 * teste, nunca senha de verdade. Todo acesso é protegido por `try` (modo privado pode negar).
 */

const SESSION_KEY = 'artifact-docs-mock-session';
const ACCOUNT_KEY = 'artifact-docs-mock-account';

type MockAccount = { name: string; email: string; password: string };

function read(key: string): string | null {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

function write(key: string, value: string | null): void {
  try {
    if (value === null) window.localStorage.removeItem(key);
    else window.localStorage.setItem(key, value);
  } catch {
    /* Sem armazenamento, a sessão dura só até a próxima recarga — e tudo bem para um mock. */
  }
}

function readAccount(): MockAccount | null {
  const raw = read(ACCOUNT_KEY);
  if (!raw) return null;

  try {
    return JSON.parse(raw) as MockAccount;
  } catch {
    return null;
  }
}

function userOf(account: MockAccount): SessionUser {
  return { id: 'user-first-access', email: account.email, name: account.name, image: null };
}

/** A sessão guarda o e-mail de quem entrou. `null` fecha a sessão. */
export function setMockSession(email: string | null): void {
  write(SESSION_KEY, email);
}

/** Apaga a conta do primeiro acesso — os testes chamam para começar do zero. */
export function resetMockAccount(): void {
  write(ACCOUNT_KEY, null);
}

export function hasMockSession(): boolean {
  return read(SESSION_KEY) !== null;
}

/** Quem está logado agora, ou 401 — como `GET /api/auth/me` no Go. */
export function mockMe(): SessionUser {
  const email = read(SESSION_KEY);
  const account = readAccount();
  if (account && email === account.email) return sessionUserSchema.parse(userOf(account));
  if (email === MOCK_USER.email) return sessionUserSchema.parse(MOCK_USER);

  throw new ApiError(401, 'Entre para continuar.');
}

export function mockSetup(): SetupStatus {
  return setupStatusSchema.parse({ needsSetup: readAccount() === null });
}

/** Confere e-mail e senha — a mesma mensagem para e-mail inexistente e senha errada. */
export function mockSignIn(body: unknown): SessionUser {
  const credentials = loginRequestSchema.safeParse(body);
  if (!credentials.success) throw new ApiError(401, INVALID_CREDENTIALS);

  const { email, password } = credentials.data;
  const account = readAccount();
  const isAccount = account?.email === email && account.password === password;
  const isTestUser = email === MOCK_USER.email && password === MOCK_PASSWORD;
  if (!isAccount && !isTestUser) throw new ApiError(401, INVALID_CREDENTIALS);

  setMockSession(email);

  return mockMe();
}

/**
 * A senha nova do link do primeiro acesso. O token é `mock:<e-mail>`: sem o prefixo, é um link
 * que não vale — 401, como o Neon Auth responde a um token vencido.
 */
export function mockSetPassword(body: unknown): void {
  const { token = '', password = '' } = (body ?? {}) as { token?: string; password?: string };
  const email = token.startsWith(MOCK_LINK_PREFIX) ? token.slice(MOCK_LINK_PREFIX.length) : '';
  if (!email) throw new ApiError(401, 'Este link não vale mais.');

  const parsed = newPasswordSchema.safeParse({ password, passwordConfirmation: password });
  if (!parsed.success) throw new ApiError(400, 'A senha precisa de pelo menos 8 caracteres.');

  write(ACCOUNT_KEY, JSON.stringify({ name: email.split('@')[0], email, password }));
}
