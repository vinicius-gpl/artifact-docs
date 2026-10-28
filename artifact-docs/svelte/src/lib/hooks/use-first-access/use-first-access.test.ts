import { beforeEach, describe, expect, it, vi } from 'vitest';

import { firstAccessErrorOf, useFirstAccessModel } from './use-first-access.svelte';

vi.mock('$lib/api/auth.api', () => ({ MOCK_LINK_PREFIX: 'mock:' }));
vi.mock('$lib/auth/neon-auth.client', () => ({ requestPasswordLink: vi.fn() }));
const mode = vi.hoisted(() => ({ mocks: false }));
vi.mock('$lib/api/http-client', () => ({
  get USE_MOCKS() {
    return mode.mocks;
  },
  readError: () => 'Não consegui falar com o servidor.',
}));

const { requestPasswordLink } = await import('$lib/auth/neon-auth.client');
const { goto } = await import('$app/navigation');

function filled(email = 'dono@exemplo.com') {
  const model = useFirstAccessModel();
  model.actions.onChange(email);

  return model;
}

beforeEach(() => {
  mode.mocks = false;
  vi.mocked(requestPasswordLink).mockReset().mockResolvedValue(undefined);
  vi.mocked(goto).mockClear();
});

describe('firstAccessErrorOf', () => {
  // feliz
  it('has nothing to say about a valid e-mail', () => {
    expect(firstAccessErrorOf('dono@exemplo.com')).toBeUndefined();
  });

  // triste
  it('asks for the e-mail when it is empty', () => {
    expect(firstAccessErrorOf('')).toBe('Informe o e-mail');
  });
});

describe('useFirstAccessModel', () => {
  // feliz
  it('asks Neon Auth for the link and says it was sent', async () => {
    const model = filled();

    model.actions.onSubmit();

    await vi.waitFor(() => expect(model.state.isSent).toBe(true));
    expect(requestPasswordLink).toHaveBeenCalledWith(
      'dono@exemplo.com',
      `${window.location.origin}/reset-password`,
    );
  });

  it('goes straight to the new password in mock mode, where there is no e-mail', async () => {
    mode.mocks = true;
    const model = filled();

    model.actions.onSubmit();

    await vi.waitFor(() =>
      expect(goto).toHaveBeenCalledWith('/reset-password?token=mock%3Adono%40exemplo.com'),
    );
    expect(requestPasswordLink).not.toHaveBeenCalled();
  });

  // triste
  it('shows the field error after trying to send, and asks for nothing', () => {
    const model = filled('dono');

    model.actions.onSubmit();

    expect(model.state.fieldError).toBe('Informe um e-mail válido');
    expect(requestPasswordLink).not.toHaveBeenCalled();
  });

  it('shows why the link was not sent, and frees the button', async () => {
    vi.mocked(requestPasswordLink).mockRejectedValue(new Error('Muitas tentativas seguidas.'));
    const model = filled();

    model.actions.onSubmit();

    await vi.waitFor(() => expect(model.state.error).toBe('Muitas tentativas seguidas.'));
    expect(model.state.isSubmitting).toBe(false);
    expect(model.state.isSent).toBe(false);
  });
});
