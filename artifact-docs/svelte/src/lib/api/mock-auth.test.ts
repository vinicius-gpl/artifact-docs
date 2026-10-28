import { beforeEach, describe, expect, it } from 'vitest';

import { MOCK_PASSWORD, MOCK_USER } from './mock-fixtures';
import {
  hasMockSession,
  mockMe,
  mockSetPassword,
  mockSetup,
  mockSignIn,
  resetMockAccount,
  setMockSession,
} from './mock-auth';

beforeEach(() => {
  resetMockAccount();
  setMockSession(null);
});

describe('first access', () => {
  // feliz
  it('gives the account a password through the link, and lets it sign in', () => {
    expect(mockSetup()).toEqual({ needsSetup: true });

    mockSetPassword({ token: 'mock:dono@exemplo.com', password: 'senha-forte' });

    expect(mockSetup()).toEqual({ needsSetup: false });
    expect(mockSignIn({ email: 'dono@exemplo.com', password: 'senha-forte' })).toMatchObject({
      email: 'dono@exemplo.com',
    });
    expect(mockMe().email).toBe('dono@exemplo.com');
  });

  // triste
  it('refuses a link that is not a mock link', () => {
    expect(() => mockSetPassword({ token: 'outro', password: 'senha-forte' })).toThrow(
      'Este link não vale mais.',
    );
    expect(mockSetup()).toEqual({ needsSetup: true });
  });

  it('refuses a password shorter than the minimum', () => {
    expect(() => mockSetPassword({ token: 'mock:dono@exemplo.com', password: 'curta' })).toThrow();
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
