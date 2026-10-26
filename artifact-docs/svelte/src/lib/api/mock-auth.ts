import {
  firstAccessSchema,
  INVALID_CREDENTIALS,
  loginRequestSchema,
  setupStatusSchema,
  type SetupStatus,
} from '$lib/contracts/auth.schema';
import { sessionUserSchema, type SessionUser } from '$lib/contracts/user.schema';

import { ApiError } from './http-client';
import { MOCK_PASSWORD, MOCK_USER } from './mock-fixtures';

/**
 * A IDENTIDADE DE MENTIRA do modo mock — o que o Neon Auth e o Go fazem em produção, imitado
 * no navegador para a tela de login funcionar sem conta nenhuma.
 *
 * Duas portas de entrada:
 * - o USUÁRIO DE TESTE (`MOCK_USER`), que existe sempre: em desenvolvimento ninguém precisa
 *   criar conta para abrir o hub;
 * - a CONTA DO PRIMEIRO ACESSO, criada pela própria tela. Enquanto ela não existe, `/auth/setup`
 *   diz `needsSetup: true` e o botão "Primeiro acesso" aparece — é assim que esse caminho é
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

/** O primeiro acesso: só enquanto não há conta — depois disso, 409, como o Go deve fazer. */
export function mockSignUp(body: unknown): SessionUser {
  if (readAccount()) {
    throw new ApiError(409, 'A conta do hub já existe. Entre com e-mail e senha.');
  }

  const account = firstAccessSchema.safeParse(body);
  if (!account.success) throw new ApiError(400, 'Confira os campos do primeiro acesso.');

  const { name, email, password } = account.data;
  write(ACCOUNT_KEY, JSON.stringify({ name, email, password }));
  setMockSession(email);

  return mockMe();
}
