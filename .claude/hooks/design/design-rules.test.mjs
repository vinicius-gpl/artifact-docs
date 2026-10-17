import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { RULES, compareWithBaseline, countByKey, violationsTouching } from './design-rules.mjs';

/** @param {string} ruleId @param {import('./design-rules.mjs').ProjectFile[]} files */
function run(ruleId, files) {
  const rule = RULES.find((candidate) => candidate.id === ruleId);
  assert.ok(rule, `unknown rule ${ruleId}`);

  return rule.check(files).map((violation) => `${violation.file} ${violation.detail}`);
}

/** @param {string} path @param {string} [content] */
function file(path, content = '') {
  return { path, content };
}

const LIB = 'svelte/src/lib/components';
const HOOKS = 'svelte/src/lib/hooks';
const ROUTES = 'svelte/src/routes';

describe('component-prefix', () => {
  // feliz
  it('accepts acerola-* folders in lib and in feature routes, in both apps', () => {
    const files = [
      file(`${LIB}/acerola-button/acerola-button.svelte`),
      file(
        `${ROUTES}/(app)/tickets/components/acerola-ticket-card/acerola-ticket-card.svelte`,
      ),
      file(`${LIB}/ui/button/button.svelte`),
      file(`${LIB}/acerola-badge/acerola-badge.svelte`),
      file(`${LIB}/ui/badge/badge.svelte`),
    ];

    assert.deepEqual(run('component-prefix', files), []);
  });

  // triste
  it('flags folders without the prefix in either app', () => {
    const files = [
      file(`${LIB}/text-field/text-field.svelte`),
      file(`${ROUTES}/(app)/dashboard/components/maintenance-log/maintenance-log.svelte`),
      file(`${LIB}/badge/badge.svelte`),
    ];

    assert.equal(run('component-prefix', files).length, 3);
  });
});

describe('feature-component-in-lib', () => {
  // triste
  it('follows the import chain to the single feature that owns it', () => {
    const files = [
      file(
        `${LIB}/acerola-dashboard-view/acerola-dashboard-view.svelte`,
        `import X from '$lib/components/acerola-peaking/acerola-peaking.svelte';`,
      ),
      file(`${LIB}/acerola-peaking/acerola-peaking.svelte`),
      file(
        `${ROUTES}/(app)/dashboard/+page.svelte`,
        `import V from '$lib/components/acerola-dashboard-view/acerola-dashboard-view.svelte';`,
      ),
    ];

    assert.deepEqual(run('feature-component-in-lib', files).sort(), [
      `${LIB}/acerola-dashboard-view só usado por routes/dashboard`,
      `${LIB}/acerola-peaking só usado por routes/dashboard`,
    ]);
  });

  // feliz
  it('keeps components used by two features or by the root layout', () => {
    const files = [
      file(`${LIB}/acerola-button/acerola-button.svelte`),
      file(`${LIB}/acerola-shell/acerola-shell.svelte`),
      file(
        `${ROUTES}/(app)/tickets/+page.svelte`,
        `components/acerola-button/acerola-button`,
      ),
      file(
        `${ROUTES}/(app)/parts/+page.svelte`,
        `components/acerola-button/acerola-button`,
      ),
      file(`${ROUTES}/+layout.svelte`, `components/acerola-shell/acerola-shell`),
    ];

    assert.deepEqual(run('feature-component-in-lib', files), []);
  });

  // triste
  /* Sem contexto, a feature é a primeira pasta de `(app)/`: duas telas de `artifacts` são uma
     feature só, e o componente que só ela usa não é genérico. */
  it('names the feature by the first folder inside (app)', () => {
    const files = [
      file(`${LIB}/acerola-button/acerola-button.svelte`),
      file(`${LIB}/acerola-preview-card/acerola-preview-card.svelte`),
      file(
        `${ROUTES}/(app)/artifacts/[label]/+page.svelte`,
        `components/acerola-preview-card/acerola-preview-card components/acerola-button/acerola-button`,
      ),
      file(
        `${ROUTES}/(app)/artifacts/+page.svelte`,
        `components/acerola-preview-card/acerola-preview-card`,
      ),
      file(`${ROUTES}/(app)/apps/+page.svelte`, `components/acerola-button/acerola-button`),
    ];

    assert.deepEqual(run('feature-component-in-lib', files), [
      `${LIB}/acerola-preview-card só usado por routes/artifacts`,
    ]);
  });

  // triste
  it('follows the import chain in the agent app, without +page', () => {
    const files = [
      file(`${LIB}/acerola-metric-tile/acerola-metric-tile.svelte`),
      file(
        `${ROUTES}/dashboard/dashboard.svelte`,
        `import M from '$lib/components/acerola-metric-tile/acerola-metric-tile.svelte';`,
      ),
    ];

    assert.deepEqual(run('feature-component-in-lib', files), [
      `${LIB}/acerola-metric-tile só usado por routes/dashboard`,
    ]);
  });

  // feliz
  it('never lets a dashboard feature own an agent component with the same name', () => {
    const files = [
      file(`${LIB}/acerola-button/acerola-button.svelte`),
      file(`${LIB}/acerola-button/acerola-button.svelte`),
      file(
        `${ROUTES}/(app)/tickets/+page.svelte`,
        `components/acerola-button/acerola-button`,
      ),
      // O agent nunca importa o próprio botão em nenhuma rota: sem dono, então não é "de feature".
    ];

    assert.deepEqual(run('feature-component-in-lib', files), [
      `${LIB}/acerola-button só usado por routes/tickets`,
    ]);
  });

  // feliz
  it('ignores generic-by-design primitives with a single consumer, prefixed or not', () => {
    const files = [
      file(`${LIB}/radar-chart/radar-chart.svelte`),
      file(`${LIB}/acerola-chart-legend/acerola-chart-legend.svelte`),
      file(`${LIB}/acerola-dialog/acerola-dialog.ts`),
      file(
        `${ROUTES}/(app)/tickets/+page.svelte`,
        `components/radar-chart/radar-chart components/acerola-chart-legend/acerola-chart-legend components/acerola-dialog/acerola-dialog`,
      ),
      file(`${LIB}/acerola-separator/acerola-separator.svelte`),
      file(
        `${ROUTES}/dashboard/dashboard.svelte`,
        `components/acerola-separator/acerola-separator`,
      ),
    ];

    assert.deepEqual(run('feature-component-in-lib', files), []);
  });

  // triste
  it('still flags a domain component whose name only resembles a generic one', () => {
    const files = [
      file(`${LIB}/acerola-usage-chart/acerola-usage-chart.svelte`),
      file(
        `${ROUTES}/(app)/computers/+page.svelte`,
        `components/acerola-usage-chart/acerola-usage-chart`,
      ),
    ];

    assert.deepEqual(run('feature-component-in-lib', files), [
      `${LIB}/acerola-usage-chart só usado por routes/computers`,
    ]);
  });
});

describe('hook-location', () => {
  // feliz
  it('accepts use-* folders and the CLI folder', () => {
    const files = [
      file(`${HOOKS}/use-ticket-list/use-ticket-list.svelte.ts`),
      file(`${HOOKS}/ui/is-mobile.svelte.ts`),
    ];

    assert.deepEqual(run('hook-location', files), []);
  });

  // triste
  it('flags loose files and folders without use-', () => {
    const files = [
      file(`${HOOKS}/use-mobile.svelte.ts`),
      file(`${HOOKS}/mirror-store/mirror-store.svelte.ts`),
    ];

    assert.equal(run('hook-location', files).length, 2);
  });

  // feliz
  it('has nothing to flag in the agent app, which has no lib/hooks', () => {
    const files = [file(`${LIB}/acerola-button/acerola-button.svelte`)];

    assert.deepEqual(run('hook-location', files), []);
  });
});

describe('lib-folder', () => {
  // feliz
  it('accepts the mapped folders in both apps', () => {
    const files = [
      file(`${LIB.replace('/components', '')}/components/x.svelte`),
      file('svelte/src/lib/components/x.svelte'),
      file('svelte/src/lib/utils/format.ts'),
    ];

    assert.deepEqual(run('lib-folder', files), []);
  });

  // triste
  it('flags the agent debt: metrics and reporting are not in the map', () => {
    const files = [
      file('svelte/src/lib/metrics/store.svelte.ts'),
      file('svelte/src/lib/reporting/client.ts'),
    ];

    assert.deepEqual(run('lib-folder', files).sort(), [
      'svelte/src/lib/metrics pasta fora do mapa',
      'svelte/src/lib/reporting pasta fora do mapa',
    ]);
  });
});

describe('class rules', () => {
  const component = (content, lib = LIB) =>
    file(`${lib}/acerola-card/acerola-card.svelte`, `<div class="${content}"></div>`);

  // feliz
  it('accepts role radius, tokens and the shadow scale', () => {
    const files = [
      component('rounded-surface border border-border bg-card text-ink-700 shadow-xs text-xs'),
    ];

    for (const rule of ['radius-by-role', 'raw-palette', 'arbitrary-font-size', 'shadow-scale']) {
      assert.deepEqual(run(rule, files), [], rule);
    }
  });

  // triste
  it('flags size radius, raw palette, pixel fonts and off-scale shadows', () => {
    const files = [component('rounded-lg text-neutral-400 text-[11px] shadow-md')];

    assert.equal(run('radius-by-role', files).length, 1);
    assert.equal(run('raw-palette', files).length, 1);
    assert.equal(run('arbitrary-font-size', files).length, 1);
    assert.equal(run('shadow-scale', files).length, 1);
  });

  // feliz
  it('lets the mobile bottom drawer keep its shadow-2xl', () => {
    const files = [
      file(
        `${LIB}/acerola-responsive-dialog-content/acerola-responsive-dialog-content.svelte`,
        `<div class="rounded-t-3xl shadow-2xl"></div>`,
      ),
    ];

    assert.deepEqual(run('shadow-scale', files), []);
  });

  // triste
  it('flags shadow-2xl anywhere else', () => {
    assert.equal(run('shadow-scale', [component('shadow-2xl')]).length, 1);
  });

  // triste
  it('flags the same violations in the agent app', () => {
    const files = [component('rounded-2xl text-neutral-400', LIB)];

    assert.equal(run('radius-by-role', files).length, 1);
    assert.equal(run('raw-palette', files).length, 1);
  });

  // feliz
  it('ignores the shadcn CLI folder, in either app', () => {
    const files = [
      file(`${LIB}/ui/card/card.svelte`, '<div class="rounded-2xl text-neutral-400"></div>'),
      file(`${LIB}/ui/card/card.svelte`, '<div class="rounded-2xl text-neutral-400"></div>'),
    ];

    assert.deepEqual(run('radius-by-role', files), []);
    assert.deepEqual(run('raw-palette', files), []);
  });
});

describe('route rules', () => {
  // triste
  it('flags raw interactive markup and control height in a dashboard route', () => {
    const files = [
      file(`${ROUTES}/(app)/tickets/+page.svelte`, '<button class="h-8">x</button>'),
    ];

    assert.equal(run('route-markup', files).length, 1);
    assert.equal(run('route-height', files).length, 1);
  });

  // triste
  it('flags the same in an agent route, which has no +page', () => {
    const files = [file(`${ROUTES}/dashboard/dashboard.svelte`, '<input class="h-9" />')];

    assert.equal(run('route-markup', files).length, 1);
    assert.equal(run('route-height', files).length, 1);
  });

  // feliz
  it('leaves a component file alone, even inside routes/', () => {
    const files = [
      file(
        `${ROUTES}/(app)/tickets/components/acerola-x/acerola-x.svelte`,
        '<button class="h-8">x</button>',
      ),
    ];

    assert.deepEqual(run('route-markup', files), []);
  });

  // triste
  it('flags a feature importing another feature component', () => {
    const files = [
      file(
        `${ROUTES}/(app)/tickets/+page.svelte`,
        `import X from '../tickets/components/acerola-x/acerola-x.svelte';`,
      ),
    ];

    assert.equal(run('cross-feature-import', files).length, 1);
  });
});

describe('storage-folder-language', () => {
  // triste
  it('flags pt-BR bucket folders', () => {
    const files = [
      file(
        'src-go/storage/artifacts.go',
        'const artifactsFolder = "artefatos/"',
      ),
    ];

    assert.equal(run('storage-folder-language', files).length, 1);
  });

  // feliz
  it('accepts English folders', () => {
    const files = [
      file(
        'src-go/storage/artifacts.go',
        'const artifactsFolder = "artifacts/"',
      ),
    ];

    assert.deepEqual(run('storage-folder-language', files), []);
  });
});

describe('doc-file-name', () => {
  // triste
  it('flags a doc file name outside kebab-case, anywhere under docs/', () => {
    const files = [file('docs/ARQUITETURA.md')];

    assert.equal(run('doc-file-name', files).length, 1);
  });

  // feliz
  it('accepts a kebab-case doc file name', () => {
    const files = [file('docs/architecture.md')];

    assert.deepEqual(run('doc-file-name', files), []);
  });
});

describe('baseline', () => {
  const violation = { rule: 'raw-palette', file: 'a.svelte', detail: 'text-neutral-400' };

  // feliz
  it('passes when nothing exceeds the baseline', () => {
    const result = compareWithBaseline([violation], countByKey([violation]));

    assert.deepEqual(result, { added: [], fixed: [] });
  });

  // triste
  it('fails on a new occurrence even of a known key', () => {
    const result = compareWithBaseline([violation, violation], countByKey([violation]));

    assert.equal(result.added.length, 1);
  });

  // feliz
  it('reports debt that was paid', () => {
    const result = compareWithBaseline([], countByKey([violation]));

    assert.deepEqual(result.fixed, ['raw-palette|a.svelte|text-neutral-400']);
  });
});

describe('violationsTouching', () => {
  const inFolder = { rule: 'component-prefix', file: `${LIB}/text-field`, detail: 'x' };
  const elsewhere = { rule: 'raw-palette', file: `${LIB}/acerola-card/acerola-card.svelte`, detail: 'x' };

  // feliz
  it('matches the violation file itself', () => {
    assert.deepEqual(
      violationsTouching([elsewhere], `${LIB}/acerola-card/acerola-card.svelte`),
      [elsewhere],
    );
  });

  // feliz
  it('matches a file inside the violation folder (component-prefix is keyed on the folder)', () => {
    assert.deepEqual(
      violationsTouching([inFolder], `${LIB}/text-field/text-field.svelte`),
      [inFolder],
    );
  });

  // triste
  it('ignores violations from unrelated files', () => {
    assert.deepEqual(violationsTouching([elsewhere], `${LIB}/other/other.svelte`), []);
  });
});
