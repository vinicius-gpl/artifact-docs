import { beforeEach, describe, expect, it, vi } from 'vitest';

import { fieldErrorsOf, useLoginModel } from './use-login.svelte';

vi.mock('$lib/api/auth.api', () => ({
  authApi: { mockSignIn: vi.fn() },
  MOCK_CREDENTIALS: { email: 'voce@exemplo.com', password: 'artifact-docs' },
}));
vi.mock('$lib/auth/neon-auth.client', () => ({ signInWithEmail: vi.fn() }));
/* O modo é decidido no build; aqui ele é trocado por teste. */
const mode = vi.hoisted(() => ({ mocks: true }));
vi.mock('$lib/api/http-client', () => ({
  get USE_MOCKS() {
    return mode.mocks;
  },
  readError: () => 'Não consegui falar com o servidor.',
}));

const { authApi } = await import('$lib/api/auth.api');
const { signInWithEmail } = await import('$lib/auth/neon-auth.client');
const { goto } = await import('$app/navigation');

const CREDENTIALS = { email: 'voce@exemplo.com', password: 'artifact-docs' };

beforeEach(() => {
  mode.mocks = true;
  vi.mocked(authApi.mockSignIn).mockReset().mockResolvedValue({
    id: 'user-mock',
    email: 'voce@exemplo.com',
    name: 'Usuário de teste',
    image: null,
  });
  vi.mocked(signInWithEmail).mockReset().mockResolvedValue(undefined);
  vi.mocked(goto).mockClear();
});

describe('fieldErrorsOf', () => {
  // feliz
  it('has nothing to say about valid credentials', () => {
    expect(fieldErrorsOf(CREDENTIALS)).toEqual({});
  });

  // triste
  it('keeps the first error of each field', () => {
    expect(fieldErrorsOf({ email: '', password: '' })).toEqual({
      email: 'Informe o e-mail',
      password: 'Informe a senha',
    });
  });
});

describe('useLoginModel', () => {
  // feliz
  it('signs in with the test user in mock mode and goes where the person was going', async () => {
    const model = useLoginModel({ returnTo: '/artifacts', initialValues: CREDENTIALS });

    model.actions.onSubmit();

    await vi.waitFor(() => expect(goto).toHaveBeenCalledWith('/artifacts'));
    expect(authApi.mockSignIn).toHaveBeenCalledWith(CREDENTIALS);
    expect(signInWithEmail).not.toHaveBeenCalled();
  });

  it('sends the credentials to Neon Auth outside mock mode', async () => {
    mode.mocks = false;
    const model = useLoginModel({ returnTo: '/ideas' });
    model.actions.onChange('email', ' voce@exemplo.com ');
    model.actions.onChange('password', 'segredo');

    model.actions.onSubmit();

    await vi.waitFor(() =>
      expect(signInWithEmail).toHaveBeenCalledWith('voce@exemplo.com', 'segredo'),
    );
    expect(authApi.mockSignIn).not.toHaveBeenCalled();
  });

  it('comes filled with the test user in mock mode, and empty in production', () => {
    expect(useLoginModel({ returnTo: '/' }).data.values).toEqual(CREDENTIALS);

    mode.mocks = false;
    expect(useLoginModel({ returnTo: '/' }).data.values).toEqual({ email: '', password: '' });
  });

  // triste
  it('shows the field errors only after trying to send, and sends nothing', () => {
    const model = useLoginModel({ returnTo: '/', initialValues: { email: '', password: '' } });
    expect(model.state.fieldErrors).toEqual({});

    model.actions.onSubmit();

    expect(model.state.fieldErrors.email).toBe('Informe o e-mail');
    expect(authApi.mockSignIn).not.toHaveBeenCalled();
  });

  it('shows the same message for a wrong e-mail or password, and frees the button', async () => {
    vi.mocked(authApi.mockSignIn).mockRejectedValue(new Error('E-mail ou senha incorretos.'));
    const model = useLoginModel({ returnTo: '/', initialValues: CREDENTIALS });

    model.actions.onSubmit();

    await vi.waitFor(() => expect(model.state.error).toBe('E-mail ou senha incorretos.'));
    expect(model.state.isSubmitting).toBe(false);
    expect(goto).not.toHaveBeenCalled();
  });

  it('explains the denied access the guard sent back', () => {
    const model = useLoginModel({ returnTo: '/', error: 'forbidden' });

    expect(model.state.error).toMatch(/não tem acesso a este hub/);
  });
});
