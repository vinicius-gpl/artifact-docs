import { describe, expect, it } from 'vitest';

import { firstAccessSchema, loginRequestSchema } from './auth.schema';

describe('loginRequestSchema', () => {
  // feliz
  it('accepts an e-mail and a password, trimming the e-mail', () => {
    expect(loginRequestSchema.parse({ email: ' voce@exemplo.com ', password: 'x' })).toEqual({
      email: 'voce@exemplo.com',
      password: 'x',
    });
  });

  // triste
  it('names the field that is wrong', () => {
    const result = loginRequestSchema.safeParse({ email: 'voce', password: '' });

    expect(result.error?.issues.map((issue) => issue.message)).toEqual([
      'Informe um e-mail válido',
      'Informe a senha',
    ]);
  });
});

describe('firstAccessSchema', () => {
  const valid = {
    name: 'Você',
    email: 'voce@exemplo.com',
    password: 'senha-forte',
    passwordConfirmation: 'senha-forte',
  };

  // feliz
  it('accepts a name, an e-mail and a confirmed password', () => {
    expect(firstAccessSchema.safeParse(valid).success).toBe(true);
  });

  // triste
  it('points the mismatch at the confirmation field', () => {
    const result = firstAccessSchema.safeParse({ ...valid, passwordConfirmation: 'outra' });

    expect(result.error?.issues[0]).toMatchObject({
      path: ['passwordConfirmation'],
      message: 'As senhas não são iguais',
    });
  });

  it('refuses a password shorter than the Neon Auth minimum', () => {
    const result = firstAccessSchema.safeParse({
      ...valid,
      password: 'curta',
      passwordConfirmation: 'curta',
    });

    expect(result.error?.issues[0]?.message).toBe('A senha precisa de pelo menos 8 caracteres');
  });
});
