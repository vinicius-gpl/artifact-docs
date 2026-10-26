import { describe, expect, it } from 'vitest';

import { siteOf } from './site-url.util';

describe('siteOf', () => {
  // feliz
  it('splits host and origin', () => {
    expect(siteOf('https://casa.exemplo.com/painel?x=1')).toEqual({
      host: 'casa.exemplo.com',
      origin: 'https://casa.exemplo.com',
    });
  });

  // triste
  it('keeps an invalid address as it came, without origin', () => {
    expect(siteOf('casa')).toEqual({ host: 'casa', origin: null });
  });
});
