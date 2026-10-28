import { beforeEach, describe, expect, it, vi } from 'vitest';

/* O Neon Auth de verdade fala com a rede; aqui só interessa o que a tela faz com a resposta. */
const neon = vi.hoisted(() => ({ requestPasswordReset: vi.fn(), resetPassword: vi.fn() }));
vi.mock('@neondatabase/neon-js/auth', () => ({ createAuthClient: () => neon }));

const { EXPIRED_LINK, requestPasswordLink, setNewPassword } = await import('./neon-auth.client');

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({}))));
  neon.requestPasswordReset.mockReset().mockResolvedValue({ data: {}, error: null });
  neon.resetPassword.mockReset().mockResolvedValue({ data: {}, error: null });
});

describe('requestPasswordLink', () => {
  // feliz
  it('asks Neon Auth for the e-mail, coming back to the new password screen', async () => {
    await requestPasswordLink('dono@exemplo.com', 'https://hub.exemplo.com/reset-password');

    expect(neon.requestPasswordReset).toHaveBeenCalledWith({
      email: 'dono@exemplo.com',
      redirectTo: 'https://hub.exemplo.com/reset-password',
    });
  });

  // triste
  it('tells to wait when there were too many attempts', async () => {
    neon.requestPasswordReset.mockResolvedValue({ data: null, error: { status: 429 } });

    await expect(requestPasswordLink('dono@exemplo.com', '/x')).rejects.toThrow('Espere');
  });

  it('turns any other refusal into an error, never into "sent"', async () => {
    neon.requestPasswordReset.mockResolvedValue({ data: null, error: { status: 500 } });

    await expect(requestPasswordLink('dono@exemplo.com', '/x')).rejects.toThrow(
      'Não consegui falar com o serviço de login',
    );
  });
});

describe('setNewPassword', () => {
  // feliz
  it('saves the password with the token of the link', async () => {
    await setNewPassword('token-do-link', 'senha-forte');

    expect(neon.resetPassword).toHaveBeenCalledWith({
      newPassword: 'senha-forte',
      token: 'token-do-link',
    });
  });

  // triste
  it('says the link expired when Neon Auth refuses the token', async () => {
    neon.resetPassword.mockResolvedValue({ data: null, error: { status: 400 } });

    await expect(setNewPassword('vencido', 'senha-forte')).rejects.toThrow(EXPIRED_LINK);
  });
});
