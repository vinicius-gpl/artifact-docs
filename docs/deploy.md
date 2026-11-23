# Deploy

O hub roda no servidor hetzner do [config](https://github.com/vinicius-gpl/config), no Swarm dele, atrás
do Caddy de lá. Quem publica é o **Komodo**: tudo do deploy está em [`.komodo/`](../.komodo/).

## Como funciona

A cada push na `main`, o webhook do GitHub dispara a Procedure `hetzner_artifact-docs` do Komodo:

1. **Pull**: o Repo `hetzner_artifact-docs` atualiza o clone da `main` no hetzner.
2. **Build**: o Build `hetzner_artifact-docs` constrói o `artifact-docs/.config/docker/dockerfile` **no próprio
   hetzner**: a SPA (Node), o binário Go com a SPA embutida e a imagem final `distroless/static:nonroot`.
   Não existe registry: a imagem fica só no hetzner, como `artifact-docs:<commit curto>-hetzner`.
3. **Deploy**: a Stack `hetzner_artifact-docs` sobe o `.komodo/artifact-docs.compose.yml` no Swarm do hetzner, com o
   disco só de leitura e um teto de 160 MB (o Go com `GOMEMLIMIT=64MiB`). As migrations rodam ao
   subir. O `komodo.sh` confere a imagem antes e, depois, o `route.sh` grava
   `conf.d/hetzner_artifact-docs.caddy` no Caddy, valida e recarrega.

Para voltar à versão anterior: Action `hetzner_artifact-docs_rollback` (rollback nativo do Swarm).

O Swarm, a rede `edge` e o Caddy são do config: este repo só cuida do que é do app.

## No GitHub

Só dois webhooks no repositório (*Settings → Webhooks*, JSON, evento *push*, com o segredo dos
webhooks do Komodo):

| Webhook | URL |
|---|---|
| Sync (aplica o `.komodo/komodo.toml`) | `https://komodo.vinicius-gpl.com/listener/github/sync/<id da sync artifact-docs>/sync` |
| Procedure (publica o app) | `https://komodo.vinicius-gpl.com/listener/github/procedure/hetzner_artifact-docs/main` |

## Variables secretas no Komodo

| Variable | De onde vem |
|---|---|
| `ARTIFACT_DOCS_DATABASE_URL` | Neon → Connect → endpoint **pooled** |
| `ARTIFACT_DOCS_NEON_AUTH_URL` | Neon → Auth → URL do endpoint (também vai no bundle; é pública) |
| `ARTIFACT_DOCS_NEON_AUTH_JWKS_URL` | Neon → Auth → JWKS URL |
| `ARTIFACT_DOCS_AUTH_ALLOWED_EMAILS` | quem pode entrar, separados por vírgula, numa linha só |
| `ARTIFACT_DOCS_R2_ENDPOINT`, `ARTIFACT_DOCS_R2_ACCESS_KEY_ID`, `ARTIFACT_DOCS_R2_SECRET_ACCESS_KEY`, `ARTIFACT_DOCS_R2_BUCKET` | Cloudflare → R2 → Manage API tokens |
| `ARTIFACT_DOCS_LINK_SECRET` | `openssl rand -hex 32` |

O endereço público é `hub.docs.vinicius-gpl.com`, em `.komodo/artifact-docs.caddy`. Dentro do
container a API escuta na porta 3000, como em todos os projetos.

## Antes do primeiro deploy

- DNS: registro `A` de `hub.docs` para o IP do hetzner na Cloudflare, em modo "DNS only", como os outros.
- Neon Auth: adicione `https://hub.docs.vinicius-gpl.com` aos domínios confiáveis.
- No Komodo: crie a Sync `artifact-docs` (repo `vinicius-gpl/artifact-docs`, branch `main`, arquivo
  `.komodo/komodo.toml`), execute e cadastre os dois webhooks acima.
- Primeiro acesso: crie o usuário no painel da Neon (Auth → Users) e use "Primeiro acesso ou
  esqueci a senha" no login para receber o link de criar a senha (`docs/auth.md`).

## Localmente

A mesma imagem de produção:

```bash
docker build -f artifact-docs/.config/docker/dockerfile --build-arg VITE_NEON_AUTH_URL=... artifact-docs
```

Para desenvolver, continua o `mise run dev`.
