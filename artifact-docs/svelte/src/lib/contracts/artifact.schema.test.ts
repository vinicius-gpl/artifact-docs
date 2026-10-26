import { describe, expect, it } from 'vitest';

import { createMockDatabase } from '$lib/api/mock-fixtures';

import { artifactLabelSchema, artifactPatchSchema, artifactSchema } from './artifact.schema';

describe('artifactLabelSchema', () => {
  // feliz
  it('accepts a slug', () => {
    expect(artifactLabelSchema.parse('abertura-cnpj-2026')).toBe('abertura-cnpj-2026');
  });

  // triste
  it('refuses uppercase, spaces and path tricks, saying what to use', () => {
    for (const label of ['Abertura', 'abertura cnpj', '../etc', '-solto', '']) {
      expect(artifactLabelSchema.safeParse(label).success).toBe(false);
    }
    expect(artifactLabelSchema.safeParse('Abertura').error?.issues[0]?.message).toBe(
      'Use só letras minúsculas, números e hífen',
    );
  });
});

describe('artifactSchema', () => {
  // feliz
  it('accepts every mock artifact — the mock follows the contract', () => {
    for (const artifact of createMockDatabase().artifacts) {
      expect(artifactSchema.safeParse(artifact).success).toBe(true);
    }
  });

  // triste
  it('refuses an artifact with an unknown kind', () => {
    const [artifact] = createMockDatabase().artifacts;

    expect(artifactSchema.safeParse({ ...artifact, kind: 'receita' }).success).toBe(false);
  });
});

describe('artifactPatchSchema', () => {
  // feliz
  it('accepts changing only the pin', () => {
    expect(artifactPatchSchema.parse({ isPinned: true })).toEqual({ isPinned: true });
  });

  // triste
  it('refuses a status that does not exist', () => {
    expect(artifactPatchSchema.safeParse({ status: 'pausado' }).success).toBe(false);
  });
});
