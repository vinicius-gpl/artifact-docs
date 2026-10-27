import { describe, expect, it } from 'vitest';

import { appLinkInputSchema } from './app-link.schema';

describe('appLinkInputSchema', () => {
  // feliz
  it('accepts an app with a full address, trimming the spaces', () => {
    const parsed = appLinkInputSchema.parse({
      name: ' Jellyfin ',
      url: ' https://jellyfin.exemplo.com ',
      typeName: ' Self-hosted ',
    });

    expect(parsed).toEqual({
      name: 'Jellyfin',
      url: 'https://jellyfin.exemplo.com',
      typeName: 'Self-hosted',
    });
  });

  // triste
  it('asks for the full address and the type, saying how', () => {
    const noScheme = appLinkInputSchema.safeParse({ name: 'X', url: 'x.com', typeName: 'T' });
    const noType = appLinkInputSchema.safeParse({ name: 'X', url: 'https://x.com', typeName: ' ' });

    expect(noScheme.error?.issues[0]?.message).toBe(
      'Informe o endereço completo, começando com https://',
    );
    expect(noType.error?.issues[0]?.message).toBe('Informe o tipo do app');
  });
});
