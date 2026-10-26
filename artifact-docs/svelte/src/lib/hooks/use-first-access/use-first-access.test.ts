import { beforeEach, describe, expect, it, vi } from 'vitest';

import { firstAccessErrorsOf, useFirstAccessModel } from './use-first-access.svelte';

vi.mock('$lib/api/auth.api', () => ({ authApi: { mockSignUp: vi.fn() } }));
vi.mock('$lib/auth/neon-auth.client', () => ({ signUpWithEmail: vi.fn() }));
const mode = vi.hoisted(() => ({ mocks: true }));
vi.mock('$lib/api/http-client', () => ({
  get USE_MOCKS() {
    return mode.mocks;
  },
  readError: () => 'Não consegui falar com o servidor.',
}));

const { authApi } = await import('$lib/api/auth.api');
const { signUpWithEmail } = await import('$lib/auth/neon-auth.client');
const { goto } = await import('$app/navigation');

const ACCOUNT = {
  name: 'Você Mesmo',
  email: 'dono@exemplo.com',
  password: 'senha-forte',
  passwordConfirmation: 'senha-forte',
};

function filled() {
  const model = useFirstAccessModel({ returnTo: '/apps' });
  for (const [field, value] of Object.entries(ACCOUNT)) {
    model.actions.onChange(field as keyof typeof ACCOUNT, value);
  }

  return model;
}

beforeEach(() => {
  mode.mocks = true;
  vi.mocked(authApi.mockSignUp).mockReset().mockResolvedValue({
    id: 'u1',
    email: ACCOUNT.email,
    name: ACCOUNT.name,
    image: null,
  });
  vi.mocked(signUpWithEmail).mockReset().mockResolvedValue(undefined);
  vi.mocked(goto).mockClear();
});

describe('firstAccessErrorsOf', () => {
  // feliz
  it('has nothing to say about a valid account', () => {
    expect(firstAccessErrorsOf(ACCOUNT)).toEqual({});
  });

  // triste
  it('points the mismatch at the confirmation', () => {
    expect(firstAccessErrorsOf({ ...ACCOUNT, passwordConfirmation: 'outra' })).toEqual({
      passwordConfirmation: 'As senhas não são iguais',
    });
  });
});

describe('useFirstAccessModel', () => {
  // feliz
  it('creates the mock account and goes in', async () => {
    const model = filled();

    model.actions.onSubmit();

    await vi.waitFor(() => expect(goto).toHaveBeenCalledWith('/apps'));
    expect(authApi.mockSignUp).toHaveBeenCalledWith(ACCOUNT);
  });

  it('creates the account in Neon Auth outside mock mode', async () => {
    mode.mocks = false;
    const model = filled();

    model.actions.onSubmit();

    await vi.waitFor(() =>
      expect(signUpWithEmail).toHaveBeenCalledWith('Você Mesmo', 'dono@exemplo.com', 'senha-forte'),
    );
  });

  // triste
  it('shows the field errors after trying to send, and creates nothing', () => {
    const model = useFirstAccessModel({ returnTo: '/' });

    model.actions.onSubmit();

    expect(model.state.fieldErrors.name).toBe('Informe o nome');
    expect(authApi.mockSignUp).not.toHaveBeenCalled();
  });

  it('shows why the account was not created, and frees the button', async () => {
    vi.mocked(authApi.mockSignUp).mockRejectedValue(new Error('A conta do hub já existe.'));
    const model = filled();

    model.actions.onSubmit();

    await vi.waitFor(() => expect(model.state.error).toBe('A conta do hub já existe.'));
    expect(model.state.isSubmitting).toBe(false);
  });
});
