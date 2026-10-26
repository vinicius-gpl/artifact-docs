/**
 * O endereço de um site em partes: o `host` que a tela mostra e a `origin` de onde sai o
 * favicon. Endereço inválido não derruba a tela: volta como veio, sem origem.
 */
export function siteOf(url: string): { host: string; origin: string | null } {
  try {
    const parsed = new URL(url);

    return { host: parsed.host, origin: parsed.origin };
  } catch {
    return { host: url, origin: null };
  }
}
