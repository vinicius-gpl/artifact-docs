import { describe, expect, it } from 'vitest';

import { markedTab, markTab } from './active-tab.svelte';

describe('markTab', () => {
  // feliz
  it('marks the tab while the screen is open and releases it on leave', () => {
    const release = markTab('ideas');
    expect(markedTab()).toBe('ideas');

    release();
    expect(markedTab()).toBeNull();
  });

  // triste
  it('does not clear a mark that another screen made after it', () => {
    const releaseFirst = markTab('plans');
    const releaseSecond = markTab('ideas');

    releaseFirst();
    expect(markedTab()).toBe('ideas');
    releaseSecond();
  });
});
