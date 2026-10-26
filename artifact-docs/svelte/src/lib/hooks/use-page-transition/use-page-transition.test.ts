import { beforeEach, describe, expect, it, vi } from 'vitest';

import { usePageTransition } from './use-page-transition.svelte';

vi.mock('$lib/motion/motion', () => ({
  pageLeave: vi.fn().mockResolvedValue(undefined),
  pageEnter: vi.fn(),
}));

const { afterNavigate, onNavigate } = await import('$app/navigation');
const { pageEnter, pageLeave } = await import('$lib/motion/motion');

type NavigateCallback = (navigation: {
  from: { url: URL } | null;
  to: { url: URL } | null;
}) => Promise<void> | void;

function at(path: string) {
  return { url: new URL(`http://localhost:5010${path}`) };
}

function registeredNavigate(): NavigateCallback {
  return vi.mocked(onNavigate).mock.calls.at(-1)?.[0] as unknown as NavigateCallback;
}

beforeEach(() => {
  vi.mocked(onNavigate).mockClear();
  vi.mocked(afterNavigate).mockClear();
  vi.mocked(pageLeave).mockClear();
  vi.mocked(pageEnter).mockClear();
});

describe('usePageTransition', () => {
  // feliz
  it('plays the exit on the page container when the route changes', async () => {
    const main = document.createElement('main');
    usePageTransition(() => main);

    await registeredNavigate()({ from: at('/apps'), to: at('/artifacts') });

    expect(pageLeave).toHaveBeenCalledWith(main);
  });

  it('plays the entrance after the new page is in place', () => {
    const main = document.createElement('main');
    usePageTransition(() => main);

    const after = vi.mocked(afterNavigate).mock.calls.at(-1)?.[0] as unknown as () => void;
    after();

    expect(pageEnter).toHaveBeenCalledWith(main);
  });

  // triste
  it('does not animate when only the query or the hash changed', async () => {
    usePageTransition(() => document.createElement('main'));

    await registeredNavigate()({ from: at('/artifacts?tipo=ideia'), to: at('/artifacts#topo') });

    expect(pageLeave).not.toHaveBeenCalled();
  });

  it('copes with the container not being mounted yet', async () => {
    usePageTransition(() => undefined);

    await registeredNavigate()({ from: at('/apps'), to: at('/ideas') });

    expect(pageLeave).toHaveBeenCalledWith(null);
  });
});
