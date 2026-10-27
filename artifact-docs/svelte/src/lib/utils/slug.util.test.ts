import { describe, expect, it } from 'vitest';

import { slugify } from './slug.util';

describe('slugify', () => {
  // feliz
  it('turns a Portuguese title into a label', () => {
    expect(slugify('Abertura do CNPJ — Ação 2026')).toBe('abertura-do-cnpj-acao-2026');
  });

  // triste
  it('returns empty for a title with no letter or number, and never ends with a hyphen', () => {
    expect(slugify('— ! —')).toBe('');
    expect(slugify(`${'a'.repeat(79)} b`)).toBe('a'.repeat(79));
  });
});
