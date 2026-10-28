import { goto } from '$app/navigation';

import { authApi } from '$lib/api/auth.api';
import { readError, USE_MOCKS } from '$lib/api/http-client';
import { signUpWithEmail } from '$lib/auth/neon-auth.client';
import { firstAccessSchema, type FirstAccessInput } from '$lib/contracts/auth.schema';

export type FirstAccessField = keyof FirstAccessInput;

export type FirstAccessModel = {
  data: { values: FirstAccessInput };
  state: {
    isSubmitting: boolean;
    error: string | null;
    fieldErrors: Partial<Record<FirstAccessField, string>>;
  };
  actions: {
    onChange: (field: FirstAccessField, value: string) => void;
    onSubmit: () => void;
  };
};

const EMPTY: FirstAccessInput = { name: '', email: '', password: '', passwordConfirmation: '' };

/** O primeiro erro de cada campo, no formato que o `TextField` mostra embaixo dele. */
export function firstAccessErrorsOf(
  values: FirstAccessInput,
): Partial<Record<FirstAccessField, string>> {
  const result = firstAccessSchema.safeParse(values);
  if (result.success) return {};

  const errors: Partial<Record<FirstAccessField, string>> = {};
  for (const issue of result.error.issues) {
    const field = issue.path[0] as FirstAccessField;
    errors[field] ??= issue.message;
  }

  return errors;
}

/**
 * O PRIMEIRO ACESSO: cria a conta do dono do hub e já entra com ela.
 *
 * Em produção a conta nasce no **Neon Auth**, direto do navegador; no modo mock, no
 * `mock-auth`. Depois de criar, segue para `returnTo` — a guarda confirma a sessão em `/api/auth/me`.
 */
export function useFirstAccessModel(input: { returnTo: string }): FirstAccessModel {
  const values = $state<FirstAccessInput>({ ...EMPTY });
  let isSubmitted = $state(false);
  let isSubmitting = $state(false);
  let error = $state<string | null>(null);

  async function createAccount(account: FirstAccessInput) {
    if (USE_MOCKS) await authApi.mockSignUp(account);
    else await signUpWithEmail(account.name, account.email, account.password);

    await goto(input.returnTo);
  }

  return {
    get data() {
      return { values };
    },
    get state() {
      return {
        isSubmitting,
        error,
        fieldErrors: isSubmitted ? firstAccessErrorsOf(values) : {},
      };
    },
    actions: {
      onChange: (field, value) => {
        values[field] = value;
      },
      onSubmit: () => {
        isSubmitted = true;
        const parsed = firstAccessSchema.safeParse(values);
        if (!parsed.success || isSubmitting) return;

        isSubmitting = true;
        error = null;
        createAccount(parsed.data)
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
