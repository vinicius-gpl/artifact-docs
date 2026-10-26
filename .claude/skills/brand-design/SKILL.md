---
name: brand-design
description: Identidade visual de Artifact-docs — padrão Carbon, paleta Personalizada, tipografia IBM Plex Sans + Plex Mono, tokens de cor e o que é proibido (cor crua, preto e branco puros). Use ANTES de criar ou mudar qualquer tela, componente, cor, espaçamento ou estilo, e ao revisar se uma tela segue a marca.
---

# Design de Artifact-docs

Esta skill é a **fonte da verdade visual**. A skill `ux-laws` diz como a tela se comporta
(tamanho de alvo, quantidade de opções, agrupamento); aqui está **como ela parece**. Um hook
confere cada arquivo editado (§7): violação de nível erro volta para você corrigir na hora.

Antes de escrever qualquer tela, responda:

1. **A cor vem de um token?** (§2) Nunca hex, rgb ou paleta padrão do Tailwind.
2. **A forma segue o padrão Carbon?** (§3: raio por papel, borda, sombra, altura)
3. **A tela respeita as leis de UX?** (skill `ux-laws`)

---

## 1. Identidade

| Papel | Escuro | Claro | Como texto (escuro / claro) |
| ----- | ------ | ----- | --------------------------- |
| **Primária** | `#ffb59e` | `#ffa58a` | `#ffb59e` / `#d13200` |
| **Destaque** | `#f8db8c` | `#ecb00e` | `#f8db8c` / `#8c6909` |
| **Terceira** | `#f88ca8` | `#f88ca8` | `#f88ca8` / `#d90d42` |

Paleta **Personalizada** (grupo Personalizada), neutros **Grafite puro**.

> Hub pessoal

 O artifact-docs é um hub pessoal  que centraliza documentação de projetos, artefatos, planejamentos, catálogo de
  aplicações e identidade visual

Sem logo definida: a marca aparece como nome escrito, com as iniciais num selo na cor primária.

## 2. Tokens

Todas as cores da interface saem destes tokens. No Tailwind viram classes (`bg-primary`,
`text-muted-foreground`, `border-border`); no CSS, `var(--primary)`. O tema escuro é marcado por
`data-theme="dark"` no `<html>`.

| Token | Papel | Claro | Escuro |
| ----- | ----- | ----- | ------ |
| `background` | Fundo da página | `#f4f4f3` | `#121214` |
| `card` | Superfície elevada (cartão, popover, diálogo) | `#fafaf9` | `#19191c` |
| `muted` | Superfície discreta (trilho, chip, tag) | `#ebebe9` | `#202024` |
| `border` | Borda e divisória | `#dcdcd8` | `#2c2c31` |
| `foreground` | Texto principal | `#141416` | `#f2f2f0` |
| `muted-foreground` | Texto secundário, legenda | `#626267` | `#a1a1a6` |
| `sidebar` | Barra lateral e cabeçalho fixo | `#efefed` | `#141417` |
| `primary` | Cor da marca cheia: ação primária (uma por tela) | `#ffa58a` | `#ffb59e` |
| `primary-foreground` | Texto sobre a cor primária | `#121214` | `#121214` |
| `primary-strong` | Primária como texto/link (passa AA no fundo) | `#d13200` | `#ffb59e` |
| `primary-soft` | Fundo suave da primária (selo, item ativo) | `#fbebe5` | `#3e3231` |
| `accent-hero` | Destaque secundário cheio (uma coisa por vez) | `#ecb00e` | `#f8db8c` |
| `accent-hero-foreground` | Texto sobre o destaque | `#121214` | `#121214` |
| `accent-strong` | Destaque como texto (eyebrow, rótulo) | `#8c6909` | `#f8db8c` |
| `tertiary` | Terceira cor (gráficos, detalhes) | `#f88ca8` | `#f88ca8` |
| `tertiary-strong` | Terceira cor como texto | `#d90d42` | `#f88ca8` |
| `ring` | Anel de foco | `#ffa58a` | `#ffb59e` |
| `shadow` | Cor das sombras (o tom escuro da marca, nunca preto) | `#121214` | `#121214` |
| `success-soft` | Fundo do selo de sucesso | `#d1e4d7` | `#21392c` |
| `success` | Sucesso (texto/ícone) | `#126e35` | `#4ade80` |
| `warning-soft` | Fundo do selo de aviso | `#eddcce` | `#3d341d` |
| `warning` | Aviso (texto/ícone) | `#9c4808` | `#fbbf24` |
| `info-soft` | Fundo do selo de informação | `#d2dbf3` | `#242f40` |
| `info` | Informação (texto/ícone) | `#1d4ed8` | `#60a5fa` |
| `destructive-soft` | Fundo do selo de erro | `#f5d4d3` | `#3d272a` |
| `destructive` | Erro e ação destrutiva | `#bb1e1e` | `#f87171` |
| `destructive-foreground` | Texto sobre o botão destrutivo | `#f4f4f3` | `#121214` |
| `chart-1` | Série 1 de gráfico | `#ff5824` | `#ffb59e` |
| `chart-2` | Série 2 de gráfico | `#b7890b` | `#f8db8c` |
| `chart-3` | Série 3 de gráfico | `#f55780` | `#f88ca8` |
| `chart-4` | Série 4 de gráfico | `#1d4ed8` | `#60a5fa` |
| `chart-5` | Série 5 de gráfico | `#dc2626` | `#f87171` |

Regras de uso:

- **`primary` cheio é raro**: uma ação primária por tela. Link e texto na cor da marca usam
  `primary-strong`, que passa no contraste.
- **`*-soft` é fundo, `*` é texto**: selo de sucesso = `bg-success-soft text-success`.
- **Texto sobre cor cheia usa o `*-foreground` dela**, nunca uma cor escolhida a olho.
- **Cor nova entra primeiro no tokens.css**, com nome de papel (`--warning`), não de cor (`--amarelo`).

## 3. Padrão: Carbon

Cantos retos, grade rígida, camadas por tom sem sombra, botões altos com texto à esquerda e IBM Plex.

Inspirado em **Carbon Design System (IBM)** (https://carbondesignsystem.com). É uma leitura livre do guia público, não a implementação
oficial: sem logo, sem fonte proprietária e com as regras desta marca por cima (alvo mínimo de
40px e sombra sem preto puro).

- Cantos retos em botões, campos, cartões e diálogos; só as tags são em pílula.
- Separação por camada de tom (fundo, camada 1, camada 2), sem sombra e sem borda em cartão.
- Campos preenchidos com só a linha de baixo.
- Botões de 48px com o texto alinhado à esquerda e espaço à direita para o ícone.
- Movimento "produtivo": rápido e funcional. Fontes IBM Plex Sans e IBM Plex Mono.

Forma por papel (tokens `--shape-*`, no Tailwind `rounded-<papel>`):

| Papel | Raio | Classe |
|---|---|---|
| Campos | `0` | `rounded-control` |
| Botões | `0` | `rounded-button` |
| Cartões | `0` | `rounded-card` |
| Diálogos | `0` | `rounded-dialog` |
| Chips e tags | `9999px` | `rounded-chip` |
| Selos | `9999px` | `rounded-badge` |

- Controles com `48px` de altura (`h-control`, token `--control-height`).
- Campos de texto: preenchido, só a linha de baixo.
- Sombra de cartão: `shadow-card` (`none`).
- Movimento: `110ms cubic-bezier(0.2, 0, 0.38, 0.9)` (`ease-brand`, `--motion-duration`).
- Nunca raio, sombra ou altura arbitrários (`rounded-[13px]`): use o papel certo.
- Referência de implementação: `.claude/skills/brand-design/reference.css` (classes `.pv-*` da prévia).

**Acabamento Vidro**: Cartões translúcidos com desfoque (backdrop-filter) sobre manchas de cor. É efeito por cima do padrão; não muda forma
nem borda. Use com moderação e no produto inteiro, nunca numa tela só.

## 4. Tipografia

- Texto: `'IBM Plex Sans', system-ui, sans-serif`
- Títulos: `'IBM Plex Sans', system-ui, sans-serif`
- Escala do Tailwind (`text-sm`, `text-base`, `text-xl`…); nunca tamanho arbitrário em px.
- Corpo de texto com no mínimo 16px em formulário no celular (evita zoom automático).

## 5. Proibido

| Proibido                                                       | Por quê                                     | No lugar                                   |
| -------------------------------------------------------------- | ------------------------------------------- | ------------------------------------------ |
| `#000`, `#fff`, `black`, `white`, `bg-black`, `text-white`     | Regra da marca: nunca preto ou branco puros | `foreground`, `background`, `*-foreground` |
| Hex, `rgb()`, `hsl()`, `oklch()` fora do tokens.css            | A cor sai da marca e não acompanha o tema   | Token                                      |
| Paleta padrão do Tailwind (`bg-blue-500`, `text-gray-600`)     | Mesma coisa                                 | Token                                      |
| Valor arbitrário de cor, raio, sombra ou fonte (`text-[13px]`) | Foge da escala                              | Classe da escala ou token novo             |
| Mais de uma ação primária na mesma tela                        | Efeito Von Restorff                         | Uma primária, o resto secundário           |

## 6. Leis de UX

- **Lei de Hick**: O tempo para decidir cresce com o número e a complexidade das opções.
- **Lei de Fitts**: O tempo para acertar um alvo depende da distância até ele e do tamanho dele.
- **Lei de Jakob**: As pessoas passam a maior parte do tempo em outros sites e esperam que o seu funcione igual.
- **Lei de Miller**: A memória de trabalho guarda poucos itens (7 ± 2) ao mesmo tempo.
- **Efeito Von Restorff**: Entre itens parecidos, o que é diferente é o que se lembra.
- **Conectividade Uniforme**: Elementos ligados visualmente (mesmo contorno, mesma linha) são percebidos como um grupo.
- **Lei de Tesler**: Toda tarefa tem uma complexidade mínima; ela deve ficar com o sistema, não com quem usa.

O detalhe, com o que exigir e o que o verificador pega, está na skill `ux-laws`.

## 7. Checagem automática

O verificador mora em `.claude/hooks/brand/` (JavaScript puro, sem dependência, Node 22+):

```bash
node .claude/hooks/brand/check-design.mjs            # varre o projeto (CI, pre-push)
node .claude/hooks/brand/check-design.mjs --json     # o mesmo, em JSON
node --test ".claude/hooks/brand/*.test.mjs"            # testes das regras e do verificador
```

- **Hook `PostToolUse`** (`.claude/settings.json`): depois de todo `Edit`/`Write`, confere o
  arquivo. Erro → você recebe a regra, a linha e a dica, e **corrige antes de seguir**. Aviso →
  chega como contexto.
- **Hook `SessionStart`**: o resumo das regras entra no contexto de toda sessão.
- Limites e leis ligadas ficam em `.claude/hooks/brand/design.config.json`.

Caso justificado (raro) se silencia na linha de cima, **com o motivo**:

```html
<!-- design-ignore-next-line fitts/target-size: ícone dentro de um alvo de 40px -->
```

Nunca silencie para "fazer passar": corrija.

## 8. Antes de entregar

- [ ] `node .claude/hooks/brand/check-design.mjs` sem erro.
- [ ] Tela conferida no tema claro **e** no escuro.
- [ ] Uma ação primária por tela; destaque usado uma vez.
- [ ] Todo alvo de toque com 40px ou mais; todo campo com rótulo.
