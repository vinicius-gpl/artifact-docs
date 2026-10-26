import { beforeEach, describe, expect, it } from 'vitest';

import { MOCK_PASSWORD, MOCK_USER } from './mock-fixtures';
import {
  hasMockSession,
  mockMe,
  mockSetup,
  mockSignIn,
  mockSignUp,
  resetMockAccount,
  setMockSession,
} from './mock-auth';

const ACCOUNT = {
  name: 'Você Mesmo',
  email: 'dono@exemplo.com',
  password: 'senha-forte',
  passwordConfirmation: 'senha-forte',
};

beforeEach(() => {
  resetMockAccount();
  setMockSession(null);
});

describe('first access', () => {
  // feliz
  it('offers the first access while there is no account, and creates it signed in', () => {
    expect(mockSetup()).toEqual({ needsSetup: true });

    const user = mockSignUp(ACCOUNT);

    expect(user).toMatchObject({ name: 'Você Mesmo', email: 'dono@exemplo.com' });
    expect(mockSetup()).toEqual({ needsSetup: false });
    expect(mockMe().email).toBe('dono@exemplo.com');
  });

  it('lets the created account sign in again later', () => {
    mockSignUp(ACCOUNT);
    setMockSession(null);

    expect(mockSignIn({ email: ACCOUNT.email, password: ACCOUNT.password }).name).toBe(
      'Você Mesmo',
    );
  });

  // triste
  it('refuses a second first access once the account exists', () => {
    mockSignUp(ACCOUNT);

    expect(() => mockSignUp({ ...ACCOUNT, email: 'outro@exemplo.com' })).toThrow(
      'A conta do hub já existe',
    );
  });

  it('refuses a first access with mismatched passwords', () => {
    expect(() => mockSignUp({ ...ACCOUNT, passwordConfirmation: 'outra' })).toThrow();
    expect(mockSetup()).toEqual({ needsSetup: true });
  });
});

describe('sign in', () => {
  // feliz
  it('always accepts the test user', () => {
    expect(mockSignIn({ email: MOCK_USER.email, password: MOCK_PASSWORD })).toEqual(MOCK_USER);
    expect(hasMockSession()).toBe(true);
  });

  // triste
  it('answers the same message for an unknown e-mail and a wrong password', () => {
    expect(() => mockSignIn({ email: 'ninguem@exemplo.com', password: 'x' })).toThrow(
      'E-mail ou senha incorretos.',
    );
    expect(() => mockSignIn({ email: MOCK_USER.email, password: 'errada' })).toThrow(
      'E-mail ou senha incorretos.',
    );
    expect(() => mockMe()).toThrow('Entre para continuar.');
  });
});
