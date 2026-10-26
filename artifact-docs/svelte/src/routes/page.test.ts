import { isRedirect } from '@sveltejs/kit';
import { describe, expect, it } from 'vitest';

import { HOME_PATH } from '$lib/navigation/navigation';

import { load } from './+page';

describe('a raiz', () => {
  // feliz
  it('opens the first screen of the menu', () => {
    let thrown: unknown;
    try {
      load();
    } catch (error) {
      thrown = error;
    }

    expect(isRedirect(thrown)).toBe(true);
    expect(thrown).toMatchObject({ status: 307, location: HOME_PATH });
  });
});
