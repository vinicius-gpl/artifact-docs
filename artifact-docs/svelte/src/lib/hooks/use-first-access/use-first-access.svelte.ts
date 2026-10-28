import { goto } from '$app/navigation';

import { MOCK_LINK_PREFIX } from '$lib/api/auth.api';
import { readError, USE_MOCKS } from '$lib/api/http-client';
import { requestPasswordLink } from '$lib/auth/neon-auth.client';
import { firstAccessSchema } from '$lib/contracts/auth.schema';

export type FirstAccessModel = {
  data: { email: string };
  state: { isSubmitting: boolean; isSent: boolean; error: string | null; fieldError?: string };
  actions: { onChange: (value: string) => void; onSubmit: () => void };
};

/** Para onde o link do e-mail traz a pessoa de volta (precisa estar nos domínios confiáveis). */
export const NEW_PASSWORD_PATH = '/reset-password';

/** O primeiro erro do e-mail, no formato que o `TextField` mostra embaixo dele. */
export function firstAccessErrorOf(email: string): string | undefined {
  const result = firstAccessSchema.safeParse({ email });

  return result.success ? undefined : result.error.issues[0]?.message;
}

/**
 * O PRIMEIRO ACESSO: pede o e-mail com o link de criar a senha (serve também para "esqueci a
 * senha"). A conta nasce no painel da Neon, sem senha nenhuma.
 *
 * O sucesso não diz se a conta existe: a tela mostra sempre "se este e-mail tiver conta…", e o
 * Neon Auth responde igual nos dois casos. No modo mock não há e-mail: a tela vai direto para a
 * senha nova, com um token que carrega o e-mail.
 */
export function useFirstAccessModel(): FirstAccessModel {
  let email = $state('');
  let isSubmitted = $state(false);
  let isSubmitting = $state(false);
  let isSent = $state(false);
  let error = $state<string | null>(null);

  async function requestLink(address: string) {
    if (USE_MOCKS) {
      await goto(`${NEW_PASSWORD_PATH}?token=${encodeURIComponent(MOCK_LINK_PREFIX + address)}`);
      return;
    }

    await requestPasswordLink(address, `${window.location.origin}${NEW_PASSWORD_PATH}`);
    isSent = true;
  }

  return {
    get data() {
      return { email };
    },
    get state() {
      return {
        isSubmitting,
        isSent,
        error,
        fieldError: isSubmitted ? firstAccessErrorOf(email) : undefined,
      };
    },
    actions: {
      onChange: (value) => {
        email = value;
      },
      onSubmit: () => {
        isSubmitted = true;
        const parsed = firstAccessSchema.safeParse({ email });
        if (!parsed.success || isSubmitting) return;

        isSubmitting = true;
        error = null;
        requestLink(parsed.data.email)
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
