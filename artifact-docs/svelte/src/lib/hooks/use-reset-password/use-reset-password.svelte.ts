import { goto } from '$app/navigation';

import { authApi } from '$lib/api/auth.api';
import { readError, USE_MOCKS } from '$lib/api/http-client';
import { setNewPassword } from '$lib/auth/neon-auth.client';
import { newPasswordSchema, type NewPasswordInput } from '$lib/contracts/auth.schema';

export type NewPasswordField = keyof NewPasswordInput;

export type ResetPasswordModel = {
  data: { values: NewPasswordInput };
  state: {
    isSubmitting: boolean;
    isLinkValid: boolean;
    error: string | null;
    fieldErrors: Partial<Record<NewPasswordField, string>>;
  };
  actions: {
    onChange: (field: NewPasswordField, value: string) => void;
    onSubmit: () => void;
  };
};

/** O primeiro erro de cada campo, no formato que o `TextField` mostra embaixo dele. */
export function newPasswordErrorsOf(
  values: NewPasswordInput,
): Partial<Record<NewPasswordField, string>> {
  const result = newPasswordSchema.safeParse(values);
  if (result.success) return {};

  const errors: Partial<Record<NewPasswordField, string>> = {};
  for (const issue of result.error.issues) {
    const field = issue.path[0] as NewPasswordField;
    errors[field] ??= issue.message;
  }

  return errors;
}

/**
 * A SENHA NOVA: a segunda metade do primeiro acesso, aberta pelo link do e-mail.
 *
 * O `token` vem do endereço, nunca de um campo: é ele que prova que a pessoa abriu o e-mail
 * dela. Sem token, a tela já abre dizendo que o link não vale, em vez de deixar preencher dois
 * campos para recusar no fim.
 *
 * Depois de salvar, manda para o login: digitar a senha uma vez confirma, na hora, que ela
 * funciona — melhor do que descobrir amanhã.
 */
export function useResetPasswordModel(input: { token: string }): ResetPasswordModel {
  const token = input.token.trim();
  const values = $state<NewPasswordInput>({ password: '', passwordConfirmation: '' });
  let isSubmitted = $state(false);
  let isSubmitting = $state(false);
  let error = $state<string | null>(null);

  async function save(password: string) {
    if (USE_MOCKS) await authApi.mockSetPassword({ token, password });
    else await setNewPassword(token, password);

    await goto('/login');
  }

  return {
    get data() {
      return { values };
    },
    get state() {
      return {
        isSubmitting,
        isLinkValid: token !== '',
        error,
        fieldErrors: isSubmitted ? newPasswordErrorsOf(values) : {},
      };
    },
    actions: {
      onChange: (field, value) => {
        values[field] = value;
      },
      onSubmit: () => {
        isSubmitted = true;
        const parsed = newPasswordSchema.safeParse(values);
        if (!parsed.success || isSubmitting || token === '') return;

        isSubmitting = true;
        error = null;
        save(parsed.data.password)
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
