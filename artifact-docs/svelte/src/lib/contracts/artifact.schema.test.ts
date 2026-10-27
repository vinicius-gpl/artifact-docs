import { describe, expect, it } from 'vitest';

import { createMockDatabase } from '$lib/api/mock-fixtures';

import {
  artifactLabelSchema,
  artifactPatchSchema,
  artifactPublishSchema,
  artifactRestoreSchema,
  artifactSchema,
  artifactVersionPublishSchema,
  type ArtifactPublishInput,
} from './artifact.schema';

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

function publishInput(overrides: Partial<ArtifactPublishInput> = {}): ArtifactPublishInput {
  return {
    file: new File(['<h1>Guia</h1>'], 'guia.html', { type: 'text/html' }),
    title: 'Guia rápido',
    label: 'guia-rapido',
    kind: 'document',
    description: '',
    tags: 'hub, marca',
    ...overrides,
  };
}

describe('artifactPublishSchema', () => {
  // feliz
  it('accepts an html file with title, label and kind', () => {
    expect(artifactPublishSchema.safeParse(publishInput()).success).toBe(true);
  });

  // triste
  it('refuses a missing title or a file the API does not publish, saying why', () => {
    const noTitle = artifactPublishSchema.safeParse(publishInput({ title: '  ' }));
    const spreadsheet = artifactPublishSchema.safeParse(
      publishInput({ file: new File(['x'], 'planilha.xlsx') }),
    );

    expect(noTitle.error?.issues[0]?.message).toBe('Informe o título');
    expect(spreadsheet.error?.issues[0]?.message).toBe('Envie um arquivo .html, .zip, .md ou .pdf');
  });
});

describe('artifactVersionPublishSchema', () => {
  // feliz
  it('accepts a new version with only the file, keeping the title', () => {
    expect(artifactVersionPublishSchema.safeParse(publishInput({ title: '' })).success).toBe(true);
  });

  // triste
  it('still refuses an empty file', () => {
    const empty = artifactVersionPublishSchema.safeParse(
      publishInput({ file: new File([], 'vazio.html') }),
    );

    expect(empty.error?.issues[0]?.message).toBe('O arquivo está vazio');
  });
});

describe('artifactRestoreSchema', () => {
  // feliz
  it('accepts a positive version', () => {
    expect(artifactRestoreSchema.safeParse({ version: 2 }).success).toBe(true);
  });

  // triste
  it('refuses version zero', () => {
    expect(artifactRestoreSchema.safeParse({ version: 0 }).success).toBe(false);
  });
});
