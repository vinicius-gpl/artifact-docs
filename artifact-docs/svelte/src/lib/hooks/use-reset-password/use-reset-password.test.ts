import { beforeEach, describe, expect, it, vi } from 'vitest';

import { newPasswordErrorsOf, useResetPasswordModel } from './use-reset-password.svelte';

vi.mock('$lib/api/auth.api', () => ({ authApi: { mockSetPassword: vi.fn() } }));
vi.mock('$lib/auth/neon-auth.client', () => ({ setNewPassword: vi.fn() }));
const mode = vi.hoisted(() => ({ mocks: false }));
vi.mock('$lib/api/http-client', () => ({
  get USE_MOCKS() {
    return mode.mocks;
  },
  readError: () => 'Não consegui falar com o servidor.',
}));

const { authApi } = await import('$lib/api/auth.api');
const { setNewPassword } = await import('$lib/auth/neon-auth.client');
const { goto } = await import('$app/navigation');

function filled(token = 'token-do-link') {
  const model = useResetPasswordModel({ token });
  model.actions.onChange('password', 'senha-forte');
  model.actions.onChange('passwordConfirmation', 'senha-forte');

  return model;
}

beforeEach(() => {
  mode.mocks = false;
  vi.mocked(setNewPassword).mockReset().mockResolvedValue(undefined);
  vi.mocked(authApi.mockSetPassword).mockReset().mockResolvedValue(undefined);
  vi.mocked(goto).mockClear();
});

describe('newPasswordErrorsOf', () => {
  // feliz
  it('has nothing to say about a confirmed password', () => {
    expect(
      newPasswordErrorsOf({ password: 'senha-forte', passwordConfirmation: 'senha-forte' }),
    ).toEqual({});
  });

  // triste
  it('points the mismatch at the confirmation', () => {
    expect(newPasswordErrorsOf({ password: 'senha-forte', passwordConfirmation: 'outra' })).toEqual(
      { passwordConfirmation: 'As senhas não são iguais' },
    );
  });
});

describe('useResetPasswordModel', () => {
  // feliz
  it('saves the password with the token of the link and goes to the login', async () => {
    const model = filled();

    model.actions.onSubmit();

    await vi.waitFor(() => expect(goto).toHaveBeenCalledWith('/login'));
    expect(setNewPassword).toHaveBeenCalledWith('token-do-link', 'senha-forte');
  });

  it('saves in the mock account in mock mode', async () => {
    mode.mocks = true;
    const model = filled('mock:dono@exemplo.com');

    model.actions.onSubmit();

    await vi.waitFor(() =>
      expect(authApi.mockSetPassword).toHaveBeenCalledWith({
        token: 'mock:dono@exemplo.com',
        password: 'senha-forte',
      }),
    );
  });

  // triste
  it('says the link is not valid when it came without a token', () => {
    const model = filled('');

    model.actions.onSubmit();

    expect(model.state.isLinkValid).toBe(false);
    expect(setNewPassword).not.toHaveBeenCalled();
  });

  it('shows why the password was not saved, and frees the button', async () => {
    vi.mocked(setNewPassword).mockRejectedValue(new Error('Este link não vale mais.'));
    const model = filled();

    model.actions.onSubmit();

    await vi.waitFor(() => expect(model.state.error).toBe('Este link não vale mais.'));
    expect(model.state.isSubmitting).toBe(false);
    expect(goto).not.toHaveBeenCalled();
  });
});
