import { goto } from '$app/navigation';

import { authApi, MOCK_CREDENTIALS } from '$lib/api/auth.api';
import { readError, USE_MOCKS } from '$lib/api/http-client';
import { signInWithEmail } from '$lib/auth/neon-auth.client';
import { loginRequestSchema, type LoginInput } from '$lib/contracts/auth.schema';
import { FORBIDDEN_ERROR } from '$lib/navigation/login-return';

export type LoginField = keyof LoginInput;

export type LoginModel = {
  data: { values: LoginInput; isMockMode: boolean };
  state: {
    isSubmitting: boolean;
    error: string | null;
    fieldErrors: Partial<Record<LoginField, string>>;
  };
  actions: {
    onChange: (field: LoginField, value: string) => void;
    onSubmit: () => void;
  };
};

const FORBIDDEN_MESSAGE = 'Esta conta não tem acesso a este hub. Entre com a conta cadastrada.';

/** No modo mock o formulário já vem com o usuário de teste; em produção, vazio. */
function initialValues(): LoginInput {
  return USE_MOCKS ? { ...MOCK_CREDENTIALS } : { email: '', password: '' };
}

/** O primeiro erro de cada campo, no formato que o `TextField` mostra embaixo dele. */
export function fieldErrorsOf(values: LoginInput): Partial<Record<LoginField, string>> {
  const result = loginRequestSchema.safeParse(values);
  if (result.success) return {};

  const errors: Partial<Record<LoginField, string>> = {};
  for (const issue of result.error.issues) {
    const field = issue.path[0] as LoginField;
    errors[field] ??= issue.message;
  }

  return errors;
}

/**
 * O LOGIN: e-mail e senha.
 *
 * - Em produção, as credenciais vão do navegador direto para o **Neon Auth** — a nossa API não
 *   vê a senha. Depois, a guarda de rota confirma com `/api/auth/me`, onde o Go confere o
 *   token e se o e-mail está na lista permitida.
 * - No modo mock (localhost), quem confere é o `mock-server`, com o usuário de teste. Os campos
 *   já vêm preenchidos com ele: em desenvolvimento ninguém precisa decorar senha de mentira.
 *
 * O formulário valida com o MESMO schema do contrato, e só depois de tentar enviar: erro
 * vermelho antes de a pessoa terminar de digitar é bronca, não ajuda.
 */
export function useLoginModel(input: {
  returnTo: string;
  error?: string | null;
  initialValues?: LoginInput;
}): LoginModel {
  const values = $state<LoginInput>(input.initialValues ?? initialValues());
  let isSubmitted = $state(false);
  let isSubmitting = $state(false);
  let error = $state<string | null>(input.error === FORBIDDEN_ERROR ? FORBIDDEN_MESSAGE : null);

  async function signIn(credentials: LoginInput) {
    if (USE_MOCKS) await authApi.mockSignIn(credentials);
    else await signInWithEmail(credentials.email, credentials.password);

    await goto(input.returnTo);
  }

  return {
    get data() {
      return { values, isMockMode: USE_MOCKS };
    },
    get state() {
      return {
        isSubmitting,
        error,
        fieldErrors: isSubmitted ? fieldErrorsOf(values) : {},
      };
    },
    actions: {
      onChange: (field, value) => {
        values[field] = value;
      },
      onSubmit: () => {
        isSubmitted = true;
        const parsed = loginRequestSchema.safeParse(values);
        if (!parsed.success || isSubmitting) return;

        isSubmitting = true;
        error = null;
        signIn(parsed.data)
          .catch((failure: unknown) => {
            error = failure instanceof Error ? failure.message : readError(failure);
          })
          .finally(() => {
            isSubmitting = false;
          });
      },
    },
  };
}
