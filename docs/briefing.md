# Briefing — artifact-docs

Hub pessoal num lugar só: documentação de projetos, artefatos HTML, planejamentos, banco de
ideias, hub de apps e casa da marca. Funciona como a tela de Artefatos do Claude: você sobe
um HTML, ele vira uma página servida e aparece numa galeria com preview.

Primeiro para uso próprio. Quando estiver pronto, vai para o Grupo Azuos.

## Origem

Junta três coisas que hoje estão separadas:

1. **`acerola-reader/docs/web`**: a documentação de projetos.
2. **`brand/`**: um HTML com links para os itens da marca do Grupo Azuos.
3. **Artefatos do Claude**: subir um HTML e ele virar uma página servida, numa galeria.

---

## Módulos

### 1. Central de documentos e artefatos

- Você cria o HTML, dá uma **label** e sobe, pela interface ou pela API.
- **Formatos:**
  - **HTML:** principal. Pode vir solto ou em `.zip`, com assets.
  - **Markdown:** convertido **uma vez, no upload**, dentro do template da marca, e salvo como `index.html`.
  - **PDF:** guardado e servido para o visualizador nativo do navegador.
- Tudo é servido como página.
- Cada upload vira uma **versão nova**. A versão atual é um ponteiro no banco, e o rollback é trocar esse ponteiro.

### 2. Galeria

Cópia do funcionamento da tela de Artefatos do Claude:

- Cards com **preview** no topo, **título**, **cadeado** (privado) e **"Editado há X"**.
- Abas de filtro, **busca**, troca **grade/lista**, **filtros** e menu de mais opções.
- **"Crie algo novo"** com os tipos: Documentos, Apresentações, Design, Painel.
- Lateral com **Fixados** e **Recentes**. "Recentes" vem do último acesso real.

### 3. Planejamentos e banco de ideias

Planejamentos (abertura do CNPJ, ERP pessoal, custo de IAM, túnel WireGuard…) e ideias são
artefatos com um **`kind`** próprio. Assim compartilham galeria, busca, preview e versões, e cada
`kind` pode ter campos próprios (ex.: status da ideia).

### 4. Hub de apps

- Cadastro de cada app: **nome**, **link**, **logo** (opcional; sem logo, o favicon é buscado automaticamente) e **tipo**.
- Os **tipos são criados por você**, e os apps aparecem **agrupados por tipo**.
- **Status online/offline**: uma goroutine pinga cada app de tempos em tempos.

### 5. Gerador de padrão de cores

- Você sobe um **`.css` no padrão do Tailwind** com as cores base.
- O backend gera as escalas (50–950, em OKLCH).
- O `.css` resultante fica **disponível para download**.
- O resultado oficial vira os **tokens da marca**, servidos pelo hub. Os templates de MD e as skills puxam dali: mudou a cor, todas as docs novas seguem.

### 6. Skills de documentação

- Skills do Claude Code que escaneiam um projeto e o **documentam 100%**, num padrão e com a marca.
- Terminam publicando pela API (`PUT /api/artifacts/<label>`) e devolvem o link.
- Geram o **thumbnail localmente** antes de publicar (veja *Preview*).

---

## Stack

| Peça | Escolha | Por quê |
|---|---|---|
| Front | **Svelte** (SvelteKit `adapter-static`, SPA) | Build entra no binário Go via `embed`. Sem Node em produção |
| Back | **Go**, binário único | RAM mínima: a VPS tem 1 GB |
| Banco | **Neon** (Postgres) | Dados, usuários e sessões fora da VPS |
| Arquivos | **Cloudflare R2** (API S3) | Artefatos, versões, thumbs, logos e CSS gerado. Egress grátis |
| Proxy | **Caddy** | Só TLS + reverse proxy para o Go |

Orçamento de RAM: Go com uns 20–30 MB (`GOMEMLIMIT` ~64 MiB) e Caddy com uns 30–50 MB. A VPS
só guarda o binário.

### Detalhes do Go

- Postgres: `pgx` com pool pequeno (`MaxConns` 4) no endpoint **pooled** do Neon.
- S3: `minio-go`, mais enxuto que o `aws-sdk-go-v2`.
- Markdown: `goldmark`.
- Artefatos são servidos em **stream** do R2 para a resposta (`io.Copy`), sem carregar o arquivo na memória, com `ETag` e `Cache-Control`.

### Neon: cuidados

- Projeto na região mais próxima da VPS: cada query vai pela rede.
- No plano free o compute suspende quando fica ocioso, e a primeira query depois disso leva ~0,5–1 s.
- Sessão validada com cache em memória (TTL curto), para não ir ao Neon a cada clique.
- Busca com `tsvector` no dicionário `portuguese` + índice GIN. O texto é extraído no upload.

---

## Rotas

Um domínio só.

```
hub.dominio
├── /                 SPA Svelte (embed)
├── /api/...          sessão (cookie) ou token de API
└── /a/<token>/       artefato: index.html
    /a/<token>/<path> assets da versão atual
```

### Link do artefato

- **O token fica no path**, não na query string. Assim os assets relativos (`style.css`, `img/...`) resolvem para `/a/<token>/style.css`. Com `?sig=`, cada asset perderia a assinatura.
- **Não usar `md5(label)`**: quem souber a label calcula o hash e abre o artefato. As opções são:
  - **token aleatório** por artefato, guardado no banco. Dá para revogar um link sem afetar os outros. **← recomendado**
  - **HMAC-SHA256(secret, label + salt)** truncado. Para revogar, precisa de um salt por artefato no banco.

### Segurança no mesmo domínio

Todo artefato é servido com:

```
Content-Security-Policy: sandbox allow-scripts allow-popups allow-forms
```

Sem `allow-same-origin`, o navegador trata o artefato como uma origem opaca: o script dele não lê
cookie nem chama a API autenticado. Junto com o cookie de sessão
`HttpOnly; SameSite=Strict; Path=/api`, um HTML enviado por upload não consegue agir como você.

Além disso:
- **label** validada como slug (`[a-z0-9-]`);
- caminhos dentro do `.zip` validados (sem `..`, sem caminho absoluto).

---

## Preview dos cards

Um Chromium headless na VPS usaria 200–400 MB, então o thumbnail é gerado **por quem publica**:

- **Web:** o navegador renderiza o HTML num iframe oculto, captura (`modern-screenshot`) e envia o PNG junto.
- **CLI/skill:** o Chrome local tira a screenshot antes do upload.
- **Sem thumb:** o card usa `<iframe loading="lazy">` escalado. O custo fica no navegador, não na VPS.

---

## Dados

### Neon (rascunho)

| Tabela | Campos principais |
|---|---|
| `users` | id, email, password_hash (argon2) |
| `sessions` | id, user_id, expires_at |
| `api_tokens` | id, user_id, name, hash, last_used_at |
| `kinds` | id, slug, name |
| `artifacts` | id, owner_id, label, title, kind_id, format (html/md/pdf), link_token, is_private, is_pinned, current_version, last_viewed_at, created_at, updated_at, search (tsvector) |
| `artifact_versions` | artifact_id, version, size, has_thumb, created_at |
| `app_types` | id, name, position |
| `apps` | id, type_id, name, url, logo_key, status, checked_at |
| `palettes` | id, name, source_key, output_key, is_brand, created_at |

`owner_id` existe desde já, pensando na ida para a empresa.

### R2

```
artifacts/<artifact_id>/v<n>/index.html
artifacts/<artifact_id>/v<n>/<assets...>
artifacts/<artifact_id>/v<n>/thumb.png
apps/<app_id>/logo.<ext>
palettes/<palette_id>/source.css
palettes/<palette_id>/output.css
```

O upload envia todos os objetos primeiro e só depois atualiza `current_version`, o que mantém
a publicação atômica.

---

## Autenticação

- Um usuário, senha com hash **argon2**, cookie de sessão.
- **Tokens de API** para as skills publicarem.
- O **cadeado** define se o link `/a/<token>/` exige login ou abre para quem tiver o link.

---

## Estrutura do repositório (proposta)

```
server/            Go (API, serve SPA e artefatos)
server/migrations/ SQL do Neon
web/               Svelte (SPA)
deploy/            Caddyfile, unit do systemd
docs/              este briefing e decisões
.claude/skills/    skills de documentação (módulo 6)
```

Os locais de commit (`.gitmessage`) já seguem esta divisão.

---

## Em aberto

1. **Link expira** ou vale até ser revogado?
2. **Público × privado:** o link sozinho basta para quem não está logado, com o cadeado exigindo login?
3. **"Documentação dinâmica":** docs que a skill republica sozinha (ex.: a cada commit), ou páginas com dados ao vivo?
4. **Gerador de cores:** a entrada é um `@theme` do Tailwind v4 com algumas cores base, e a saída são as escalas 50–950?
5. **"Crie algo novo":** templates com a marca, atalho de upload, ou outra coisa?
6. **Abas "Meus / Compartilhados":** ficam como estão (pensando na empresa) ou viram filtros por `kind` enquanto for só você?
7. **Empresa depois:** a mesma instância com multiusuário, ou uma cópia separada?
