import type { LucideIcon } from '@lucide/svelte';
import AppWindow from '@lucide/svelte/icons/app-window';
import Lightbulb from '@lucide/svelte/icons/lightbulb';
import LayoutGrid from '@lucide/svelte/icons/layout-grid';
import ListTodo from '@lucide/svelte/icons/list-todo';
import Palette from '@lucide/svelte/icons/palette';

/**
 * O MENU LATERAL — uma lista só, porque o hub é um módulo só.
 *
 * Cada item é uma pasta em `src/routes/(app)` (`/artifacts`, `/plans`…). O item fica aceso
 * também nas subrotas: `/artifacts/briefing` acende "Artefatos".
 *
 * Tela nova no menu = uma linha aqui.
 */
export type NavItem = {
  /** Única no menu: é o endereço sem a barra inicial (`artifacts`). */
  key: string;
  label: string;
  to: string;
  icon: LucideIcon;
};

function navItem(key: string, label: string, icon: LucideIcon): NavItem {
  return { key, label, to: `/${key}`, icon };
}

/**
 * A ORDEM É A DO USO: começa no que se abre todo dia (Artefatos, Planejamentos, Ideias), passa
 * pelos atalhos (Apps) e termina na marca, que se mexe de vez em quando.
 */
export const NAV_ITEMS: readonly NavItem[] = [
  navItem('artifacts', 'Artefatos', LayoutGrid),
  navItem('plans', 'Planejamentos', ListTodo),
  navItem('ideas', 'Ideias', Lightbulb),
  navItem('apps', 'Apps', AppWindow),
  navItem('brand', 'Marca', Palette),
];

/** A tela em que o sistema abre. */
export const HOME_PATH = '/artifacts';

/**
 * O item aceso, DENTRO DOS ITENS QUE A CASCA RECEBEU — o prefixo mais longo que casar vence,
 * para `/artifacts/archive` acender "Arquivo" e não "Artefatos", se os dois existirem.
 */
export function activeNavKeyOf(pathname: string, items: readonly NavItem[]): string | undefined {
  return navItemOf(pathname, items)?.key;
}

/** O item de menu que responde por um caminho: o casamento mais específico. */
function navItemOf(pathname: string, items: readonly NavItem[]): NavItem | undefined {
  const matches = items.filter(
    (item) => pathname === item.to || pathname.startsWith(`${item.to}/`),
  );

  return matches.sort((a, b) => b.to.length - a.to.length)[0];
}
