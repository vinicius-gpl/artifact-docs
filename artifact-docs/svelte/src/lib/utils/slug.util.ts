/** O maior slug que vira label (o `artifactLabelSchema` aceita até 80). */
const MAX_SLUG_LENGTH = 80;

/**
 * Transforma um título em label: "Abertura do CNPJ — 2026" → "abertura-do-cnpj-2026".
 *
 * Tira o acento (o "ç" vira "c"), troca tudo o que não é letra ou número por hífen e não
 * deixa hífen sobrando nas pontas. É o que a tela sugere enquanto a pessoa digita o título;
 * ela continua podendo editar a label.
 */
export function slugify(text: string): string {
  return text
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .slice(0, MAX_SLUG_LENGTH)
    .replace(/^-+|-+$/g, '');
}
