import type { LucideIcon } from '@lucide/svelte';
import AppWindow from '@lucide/svelte/icons/app-window';
import FileText from '@lucide/svelte/icons/file-text';
import Lightbulb from '@lucide/svelte/icons/lightbulb';
import ListTodo from '@lucide/svelte/icons/list-todo';

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
  /** O grupo do item. A barra separa os grupos (uma linha no desktop, um título no menu do
   * celular); item sem grupo fica sozinho. */
  group?: string;
};

/** O grupo das abas de publicação — cada uma um tipo de artefato. */
export const PUBLISH_GROUP = 'Publicações';

function navItem(key: string, label: string, icon: LucideIcon, group?: string): NavItem {
  return { key, label, to: `/${key}`, icon, group };
}

/** Diz se o item abre um grupo novo — é antes dele que a barra põe a separação. */
export function startsGroup(items: readonly NavItem[], index: number): boolean {
  return index > 0 && items[index]?.group !== items[index - 1]?.group;
}

/**
 * A ORDEM É A DO USO: começa nos Apps, a página inicial — o hub é antes de tudo o ponto de
 * partida para os outros apps —, e segue pelo que se publica.
 *
 * Cada aba de publicação é UM tipo de artefato (Documentos, Planejamentos, Ideias): mostra e
 * publica só o seu. Uma aba "de tudo" ao lado das outras deixava a pessoa sem saber onde um
 * planejamento aparece e por onde publicá-lo. O endereço de Documentos continua `/artifacts`,
 * que é também onde mora o detalhe de qualquer artefato (`/artifacts/<label>`).
 */
export const NAV_ITEMS: readonly NavItem[] = [
  navItem('apps', 'Apps', AppWindow),
  navItem('artifacts', 'Documentos', FileText, PUBLISH_GROUP),
  navItem('plans', 'Planejamentos', ListTodo, PUBLISH_GROUP),
  navItem('ideas', 'Ideias', Lightbulb, PUBLISH_GROUP),
];

/** A tela em que o sistema abre: a primeira do menu. */
export const HOME_PATH = '/apps';

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
