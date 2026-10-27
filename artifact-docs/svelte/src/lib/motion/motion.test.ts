import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import {
  flipFrom,
  hoverLift,
  moveIndicator,
  pageEnter,
  pageLeave,
  prefersReducedMotion,
  revealChildren,
} from './motion';

/* O GSAP de verdade anima no relógio do navegador; aqui só interessa O QUE foi pedido a ele.
   `to` chama o `onComplete` na hora, para a promessa da saída de página resolver. */
vi.mock('gsap', () => ({
  default: {
    to: vi.fn((_target: unknown, vars: { onComplete?: () => void }) => vars.onComplete?.()),
    set: vi.fn(),
    fromTo: vi.fn(),
    from: vi.fn(),
  },
}));

const gsap = (await import('gsap')).default;

function setReducedMotion(reduce: boolean, hover = true) {
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: query.includes('reduce') ? reduce : hover,
    })),
  );
}

function pageWithBlocks(count: number): HTMLElement {
  const main = document.createElement('main');
  const page = document.createElement('div');
  for (let index = 0; index < count; index++) page.append(document.createElement('section'));
  main.append(page);

  return main;
}

beforeEach(() => {
  vi.mocked(gsap.to).mockClear();
  vi.mocked(gsap.set).mockClear();
  vi.mocked(gsap.fromTo).mockClear();
  setReducedMotion(false);
});

afterEach(() => {
  vi.unstubAllGlobals();
  delete document.documentElement.dataset.effects;
});

describe('prefersReducedMotion', () => {
  // feliz
  it('follows the operating system preference', () => {
    setReducedMotion(true);

    expect(prefersReducedMotion()).toBe(true);
  });

  // triste
  it('also turns motion off in the lite effects mode', () => {
    document.documentElement.dataset.effects = 'lite';

    expect(prefersReducedMotion()).toBe(true);
  });
});

describe('pageLeave and pageEnter', () => {
  // feliz
  it('fades the leaving page and resolves when it is done', async () => {
    const main = document.createElement('main');

    await pageLeave(main);

    expect(gsap.to).toHaveBeenCalledWith(main, expect.objectContaining({ opacity: 0 }));
  });

  it('brings the blocks of the new page in, one after the other', () => {
    const main = pageWithBlocks(3);

    pageEnter(main);

    const [targets, , vars] = vi.mocked(gsap.fromTo).mock.calls[0]!;
    expect(targets).toHaveLength(3);
    expect(vars).toMatchObject({ opacity: 1, y: 0, stagger: expect.any(Number) });
  });

  // triste
  it('does not hold the navigation when motion is reduced', async () => {
    setReducedMotion(true);

    await pageLeave(document.createElement('main'));

    expect(gsap.to).not.toHaveBeenCalled();
  });

  it('still makes the new page visible when motion is reduced', () => {
    setReducedMotion(true);
    const main = pageWithBlocks(2);

    pageEnter(main);

    expect(gsap.set).toHaveBeenCalledWith(main, expect.objectContaining({ opacity: 1 }));
    expect(gsap.fromTo).not.toHaveBeenCalled();
  });
});

describe('revealChildren', () => {
  // feliz
  it('staggers the children in, with a cap on the total delay', () => {
    const grid = pageWithBlocks(5).firstElementChild as HTMLElement;

    revealChildren(grid);

    expect(vi.mocked(gsap.fromTo).mock.calls[0]?.[2]).toMatchObject({
      stagger: { amount: expect.any(Number) },
    });
  });

  // triste
  it('does nothing on an empty list (edge case)', () => {
    revealChildren(document.createElement('div'));

    expect(gsap.fromTo).not.toHaveBeenCalled();
  });
});

describe('moveIndicator', () => {
  // feliz
  it('slides the marker to the active item', () => {
    const indicator = document.createElement('span');
    const active = document.createElement('a');

    moveIndicator(indicator, active);

    expect(gsap.to).toHaveBeenCalledWith(indicator, expect.objectContaining({ opacity: 1 }));
  });

  it('only places the marker on the first measure, without travelling', () => {
    const indicator = document.createElement('span');

    moveIndicator(indicator, document.createElement('a'), { immediate: true });

    expect(gsap.set).toHaveBeenCalled();
    expect(gsap.to).not.toHaveBeenCalled();
  });

  // triste
  it('hides the marker when no item is active', () => {
    const indicator = document.createElement('span');

    moveIndicator(indicator, null);

    expect(gsap.to).toHaveBeenCalledWith(indicator, expect.objectContaining({ opacity: 0 }));
  });
});

describe('hoverLift', () => {
  // feliz
  it('lifts the card under the mouse and settles it back', () => {
    const card = document.createElement('article');
    hoverLift(card);

    card.dispatchEvent(new Event('pointerenter'));
    card.dispatchEvent(new Event('pointerleave'));

    expect(vi.mocked(gsap.to).mock.calls.map((call) => (call[1] as { y: number }).y)).toEqual([
      -4, 0,
    ]);
  });

  // triste
  it('does nothing on a touch screen, where there is no hover', () => {
    setReducedMotion(false, false);
    const card = document.createElement('article');
    hoverLift(card);

    card.dispatchEvent(new Event('pointerenter'));

    expect(gsap.to).not.toHaveBeenCalled();
  });
});

describe('flipFrom', () => {
  function placed(rect: Partial<DOMRect>): HTMLElement {
    const node = document.createElement('section');
    node.getBoundingClientRect = () =>
      ({ left: 0, top: 0, width: 0, height: 0, ...rect }) as DOMRect;

    return node;
  }

  // feliz
  it('animates from the old box back to the new one, only with transform', () => {
    setReducedMotion(false);
    vi.mocked(gsap.fromTo).mockClear();
    const node = placed({ left: 0, top: 0, width: 1000, height: 800 });

    flipFrom(node, { left: 100, top: 200, width: 500, height: 400 } as DOMRect);

    expect(gsap.fromTo).toHaveBeenCalledWith(
      node,
      expect.objectContaining({ x: 100, y: 200, scaleX: 0.5, scaleY: 0.5 }),
      expect.objectContaining({ x: 0, y: 0, scaleX: 1, scaleY: 1 }),
    );
  });

  // triste
  it('does not animate when the person asks for less motion, or the element has no size', () => {
    vi.mocked(gsap.fromTo).mockClear();
    setReducedMotion(true);
    flipFrom(placed({ width: 1000, height: 800 }), {
      left: 0,
      top: 0,
      width: 1,
      height: 1,
    } as DOMRect);

    setReducedMotion(false);
    flipFrom(placed({}), { left: 0, top: 0, width: 1, height: 1 } as DOMRect);

    expect(gsap.fromTo).not.toHaveBeenCalled();
  });
});
