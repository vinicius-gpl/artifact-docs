/**
 * A ABA ACESA QUANDO O ENDEREÇO NÃO DIZ QUAL É.
 *
 * O detalhe de qualquer artefato mora em `/artifacts/<label>` — e `/artifacts` é a aba
 * Documentos. Uma ideia aberta acenderia "Documentos". Quem sabe a aba certa é o detalhe,
 * depois de carregar o artefato: ele marca a aba do tipo enquanto está na tela e solta ao sair.
 *
 * Só a tela marca, só a barra lê (`use-top-nav`).
 */
const marked = $state<{ key: string | null }>({ key: null });

/** A aba marcada pela tela aberta, ou `null` quando a barra deve seguir o endereço. */
export function markedTab(): string | null {
  return marked.key;
}

/** Marca a aba e devolve a função que solta a marca — o retorno de um `$effect`. */
export function markTab(key: string): () => void {
  marked.key = key;

  return () => {
    if (marked.key === key) marked.key = null;
  };
}
