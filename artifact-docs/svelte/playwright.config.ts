import { readFileSync } from 'node:fs';

import { defineConfig, devices } from '@playwright/test';

/**
 * Lê uma variável do `src-go/.env` sem depender do `dotenv` (o client não tem essa lib): são
 * poucas, e todas já vêm documentadas em `src-go/.env.example`.
 */
function readServerEnv(name: string): string | undefined {
  if (process.env[name]) return process.env[name];
  try {
    const envFile = readFileSync('../src-go/.env', 'utf8');

    return envFile.match(new RegExp(`^${name}=(.*)$`, 'm'))?.[1]?.trim() || undefined;
  } catch {
    return undefined;
  }
}

const testDatabaseUrl = readServerEnv('TEST_DATABASE_URL');
if (testDatabaseUrl) process.env.TEST_DATABASE_URL = testDatabaseUrl;

/**
 * E2E da WEB: o navegador de verdade, a API de verdade e um banco descartável.
 *
 * O Playwright sobe o `mise run dev` da raiz sozinho (`webServer`) e aponta o server para a
 * `TEST_DATABASE_URL` (uma branch de banco só para teste na Neon) — nunca a `DATABASE_URL` de
 * quem está desenvolvendo, que perderia os dados a cada execução.
 *
 * Sem `TEST_DATABASE_URL` no `src-go/.env`, o teste em si é pulado (ver os testes em `e2e/`) em
 * vez de arriscar rodar contra o banco de trabalho. O hub não tem login: o teste abre as telas
 * direto.
 *
 * Na primeira vez, baixe o navegador:  npx playwright install chromium
 */
export default defineConfig({
  testDir: './e2e',
  testMatch: '**/*.e2e.ts',
  fullyParallel: false,
  retries: process.env.CI ? 1 : 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  /**
   * Quinze segundos, e não os cinco do padrão.
   *
   * Em desenvolvimento o Vite compila os módulos de cada rota NA PRIMEIRA VEZ que alguém a
   * abre. A primeira navegação de uma execução leva alguns segundos antes de a tela reagir, e
   * com cinco segundos a asserção estourava ali — parecendo defeito da tela, que está certa.
   *
   * Aumentar o teto não deixa teste ruim passar: quem falha de verdade falha igual, só demora
   * mais para desistir.
   */
  expect: { timeout: 15_000 },
  use: {
    baseURL: 'http://localhost:5010',
    locale: 'pt-BR',
    trace: 'retain-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: 'mise run dev',
    cwd: '..',
    /**
     * ESPERA A API, não a tela.
     *
     * A tela do Vite sobe em segundos; a API precisa compilar o Go e falar com o banco antes
     * de responder. Esperando só a tela, o teste começa enquanto a
     * API ainda compila e TODA requisição volta 502 — o sintoma parece defeito da tela.
     *
     * A rota de saúde é o sinal certo porque ela só responde 200 depois de a API falar com o
     * banco. E esperar pela API já implica a tela: ela subiu muito antes.
     */
    url: `http://localhost:${readServerEnv('API_PORT') ?? '3010'}/api/health`,
    reuseExistingServer: !process.env.CI,
    /* Generoso porque a conta é a compilação fria do server, não a partida dele. */
    timeout: 240_000,
    env: testDatabaseUrl ? { DATABASE_URL: testDatabaseUrl } : {},
  },
});
