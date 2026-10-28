import gsap from 'gsap';

/**
 * As constantes de movimento do sistema, num lugar só.
 *
 * Duração e curva são identidade, do mesmo jeito que cor e tipografia: quando cada tela
 * escolhe a sua, a mesma transição fica de um jeito na lista e de outro no painel, e o
 * sistema passa a parecer montado por pessoas diferentes — que é o que ele não deve parecer.
 *
 * Os valores são curtos de propósito. Animação em ferramenta de trabalho serve para explicar
 * o que mudou de lugar, não para ser apreciada: acima de ~350ms ela vira espera, e uma
 * pessoa que abre esta tela quarenta vezes por dia sente cada uma delas.
 */
export const DURATION = {
  /** Reação a um clique — abrir, fechar, marcar. */
  fast: 0.18,
  /** Entrada de conteúdo. */
  base: 0.32,
  /** Números que sobem. Mais longo porque o olho precisa acompanhar a contagem. */
  count: 0.7,
} as const;

export const EASE = {
  /** Sai rápido e freia — a curva de algo que chega e para. */
  out: 'power2.out',
  /** Para o que abre e fecha; dá peso ao movimento sem exagero. */
  inOut: 'power2.inOut',
} as const;

/**
 * A PESSOA MANDA MAIS QUE O DESIGN.
 *
 * Quem liga "reduzir movimento" no sistema operacional costuma fazê-lo por enjoo,
 * enxaqueca ou vertigem — não por preferência estética. Aqui isso não vira "animação mais
 * curta": vira nenhuma. Toda função deste arquivo respeita esta checagem, e por isso ela
 * mora aqui e não em cada componente, onde alguém esqueceria.
 */
export function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined' || !window.matchMedia) return false;
  /* O modo leve (`data-effects='lite'`, decidido por `use-effects`) também desliga: a máquina
     não dá conta, e animação travada é pior que nenhuma. */
  if (document.documentElement.dataset.effects === 'lite') return true;

  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/** Entrada padrão: sobe alguns pixels enquanto aparece. */
export function fadeInUp(targets: gsap.TweenTarget, options: { delay?: number } = {}): void {
  if (prefersReducedMotion()) return;

  gsap.from(targets, {
    opacity: 0,
    y: 12,
    duration: DURATION.base,
    ease: EASE.out,
    delay: options.delay ?? 0,
    clearProps: 'transform',
  });
}

/**
 * Entrada em cascata, para listas e grupos de cartões.
 *
 * `stagger` pequeno de propósito: o atraso existe para dar ordem de leitura, não para fazer
 * a última linha chegar meio segundo depois da primeira.
 */
export function staggerIn(targets: gsap.TweenTarget, options: { delay?: number } = {}): void {
  if (prefersReducedMotion()) return;

  gsap.from(targets, {
    opacity: 0,
    y: 10,
    duration: DURATION.base,
    ease: EASE.out,
    stagger: 0.05,
    delay: options.delay ?? 0,
  });
}

/**
 * Item que ACABOU de virar o ativo — um pulso rápido, não um efeito chamativo.
 *
 * `fromTo` (não `from`): parte de um valor conhecido (`0.7`) e sempre TERMINA em `1`, escrito
 * explicitamente. Só `from` deixaria o estado final por conta da última leitura do CSS — e
 * clicar duas vezes seguidas no mesmo item, cortando a animação no meio, arriscava travar o
 * ícone menor que o normal para sempre.
 */
export function popIn(targets: gsap.TweenTarget, options: { delay?: number } = {}): void {
  if (prefersReducedMotion()) return;

  gsap.fromTo(
    targets,
    { scale: 0.7 },
    { scale: 1, duration: DURATION.fast, ease: EASE.out, delay: options.delay ?? 0 },
  );
}

/**
 * O FUNDO de "isto é o item ativo agora" — nasce com um fade, não troca seca de cor.
 *
 * `color` chega resolvido (não `"var(--token)"` cru): o motor de cor do GSAP não sabe
 * interpolar uma variável CSS — sem valor de verdade pra partir, ele só troca no talo no
 * fim da animação, e o fade vira o mesmo corte seco que essa função existe pra tirar.
 * `clearProps` no fim devolve o controle pra classe do Tailwind (`data-[active=true]:bg-*`),
 * pra um hover ou uma troca de tema logo depois não ficar preso no valor fixo desta tween.
 */
export function highlightIn(
  target: Element | null,
  options: { color?: string; delay?: number } = {},
): void {
  if (!target) return;

  const color =
    options.color ??
    getComputedStyle(document.documentElement).getPropertyValue('--sidebar-accent').trim();

  if (prefersReducedMotion()) return;

  gsap.fromTo(
    target,
    { backgroundColor: 'transparent' },
    {
      backgroundColor: color,
      duration: DURATION.fast,
      ease: EASE.out,
      delay: options.delay ?? 0,
      clearProps: 'backgroundColor',
    },
  );
}

/**
 * Número que sobe até o valor.
 *
 * O valor final é escrito no elemento ANTES de animar, e de novo no fim: se a animação for
 * cortada no meio — troca de rota, movimento reduzido, aba em segundo plano — o que fica na
 * tela é o número certo, nunca um valor parcial. Painel que mostra "13" quando são 23 é pior
 * que painel sem animação nenhuma.
 */
export function countTo(element: HTMLElement | null, value: number): void {
  if (!element) return;

  if (prefersReducedMotion()) {
    element.textContent = String(value);

    return;
  }

  const counter = { current: 0 };
  gsap.to(counter, {
    current: value,
    duration: DURATION.count,
    ease: EASE.out,
    onUpdate: () => {
      element.textContent = String(Math.round(counter.current));
    },
    onComplete: () => {
      element.textContent = String(value);
    },
  });
}

/* ==========================================================================
   A TROCA DE PÁGINA e as entradas — o que a navegação do hub anima.
   ========================================================================== */

/**
 * A tela que está SAINDO: some e sobe um pouco, rápido. A navegação espera esta promessa
 * (`onNavigate`), então ela tem de ser curta — é tempo em que a pessoa já clicou e ainda não
 * vê a tela nova. Com movimento reduzido, resolve na hora.
 */
export function pageLeave(target: Element | null): Promise<void> {
  if (!target || prefersReducedMotion()) return Promise.resolve();

  return new Promise((resolve) => {
    gsap.to(target, {
      opacity: 0,
      y: -8,
      duration: DURATION.fast * 0.8,
      ease: 'power1.in',
      onComplete: resolve,
    });
  });
}

/**
 * A tela que está CHEGANDO: o contêiner volta ao normal e os blocos da tela (título, filtros,
 * conteúdo) entram em cascata, de cima para baixo — a ordem em que se lê.
 *
 * `clearProps` no fim de tudo: transform e opacidade voltam para o CSS. Sem isso, um
 * `transform` esquecido no contêiner cria um novo contexto de empilhamento e o `sticky` e os
 * balões de dentro da tela passam a se comportar errado.
 */
export function pageEnter(target: HTMLElement | null): void {
  if (!target) return;

  gsap.set(target, { opacity: 1, y: 0, clearProps: 'opacity,transform' });
  if (prefersReducedMotion()) return;

  const blocks = target.firstElementChild?.children ?? target.children;
  gsap.fromTo(
    blocks,
    { opacity: 0, y: 18 },
    {
      opacity: 1,
      y: 0,
      duration: DURATION.base + 0.08,
      ease: 'power3.out',
      stagger: 0.07,
      clearProps: 'opacity,transform',
    },
  );
}

/**
 * Action: os FILHOS do elemento entram em cascata quando ele aparece — a grade de cartões, a
 * lista de grupos. `stagger` com teto (`amount`): numa grade de 40 cartões, o último não pode
 * chegar dois segundos depois do primeiro.
 */
export function revealChildren(node: HTMLElement): void {
  if (prefersReducedMotion() || node.children.length === 0) return;

  gsap.fromTo(
    node.children,
    { opacity: 0, y: 14, scale: 0.98 },
    {
      opacity: 1,
      y: 0,
      scale: 1,
      duration: DURATION.base,
      ease: 'power2.out',
      stagger: { each: 0.045, amount: 0.4 },
      clearProps: 'opacity,transform',
    },
  );
}

/** Action: o bloco abre DESLIZANDO de cima (o menu do celular), em vez de surgir de uma vez. */
export function slideDown(node: HTMLElement): void {
  if (prefersReducedMotion()) return;

  gsap.fromTo(
    node,
    { height: 0, opacity: 0 },
    {
      height: 'auto',
      opacity: 1,
      duration: DURATION.base,
      ease: EASE.out,
      clearProps: 'height,opacity',
    },
  );
  gsap.fromTo(
    node.querySelectorAll('li'),
    { opacity: 0, x: -12 },
    { opacity: 1, x: 0, duration: DURATION.base, ease: EASE.out, stagger: 0.04, delay: 0.05 },
  );
}

/**
 * O marcador da navbar DESLIZA até o item aceso: mede o item dentro da lista e anima posição e
 * largura. Na primeira medida (`immediate`) ele só se posiciona — a página que abre já vem
 * com o marcador no lugar, sem viajar do canto da tela.
 */
export function moveIndicator(
  indicator: HTMLElement | null,
  active: HTMLElement | null,
  options: { immediate?: boolean } = {},
): void {
  if (!indicator) return;
  if (!active) {
    gsap.to(indicator, { opacity: 0, duration: DURATION.fast });

    return;
  }

  const props = { x: active.offsetLeft, width: active.offsetWidth, opacity: 1 };
  if (options.immediate || prefersReducedMotion()) {
    gsap.set(indicator, props);

    return;
  }

  gsap.to(indicator, { ...props, duration: DURATION.base, ease: 'power3.out' });
}

/**
 * Action: o cartão LEVANTA um pouco sob o mouse e volta ao sair. Só onde existe mouse de
 * verdade — no toque não há hover, e o cartão ficaria levantado depois do toque.
 */
export function hoverLift(node: HTMLElement): { destroy: () => void } {
  const hasMouse = window.matchMedia?.('(hover: hover) and (pointer: fine)').matches ?? false;
  if (!hasMouse) return { destroy: () => {} };

  const lift = () => {
    if (prefersReducedMotion()) return;
    gsap.to(node, { y: -4, duration: DURATION.fast, ease: EASE.out });
  };
  const settle = () => {
    gsap.to(node, { y: 0, duration: DURATION.base, ease: EASE.out, clearProps: 'transform' });
  };

  node.addEventListener('pointerenter', lift);
  node.addEventListener('pointerleave', settle);

  return {
    destroy: () => {
      node.removeEventListener('pointerenter', lift);
      node.removeEventListener('pointerleave', settle);
    },
  };
}

/**
 * O ELEMENTO MUDOU DE LUGAR E DE TAMANHO — anima a mudança em vez de saltar (a técnica FLIP).
 *
 * Quem chama mede o retângulo ANTES (`first`), aplica a mudança (ex.: a prévia vira tela
 * cheia) e espera o DOM atualizar; aqui o elemento já está no lugar novo. Ele volta, só com
 * `transform`, para onde estava, e anda até o lugar novo — o olho acompanha a prévia crescendo
 * até cobrir a janela, ou encolhendo de volta para o lugar dela.
 *
 * Só `transform`: nada é recalculado no layout durante a animação, que fica no compositor.
 *
 * Devolve uma promessa que resolve quando a animação termina (na hora, sem animação): quem
 * chama precisa saber quando tirar o que segura o elemento por cima do resto durante o voo.
 * `overwrite` derruba uma animação anterior ainda em curso — clicar duas vezes rápido não
 * deixa duas brigando pelo mesmo `transform`.
 */
export function flipFrom(node: HTMLElement, first: DOMRect): Promise<void> {
  if (prefersReducedMotion()) return Promise.resolve();

  const last = node.getBoundingClientRect();
  if (!last.width || !last.height) return Promise.resolve();

  return new Promise((resolve) => {
    gsap.fromTo(
      node,
      {
        x: first.left - last.left,
        y: first.top - last.top,
        scaleX: first.width / last.width,
        scaleY: first.height / last.height,
        transformOrigin: 'top left',
      },
      {
        x: 0,
        y: 0,
        scaleX: 1,
        scaleY: 1,
        duration: DURATION.base,
        ease: EASE.inOut,
        overwrite: true,
        clearProps: 'transform,transformOrigin',
        onComplete: () => resolve(),
      },
    );
  });
}
