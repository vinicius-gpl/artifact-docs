# Deploy

O hub roda na VPS Oracle do [my-config](https://github.com/vinicius-gpl/my-config), atrás do Caddy
de lá. Push na `main` que mexe no app ou no deploy dispara o workflow **Deploy**
(`.github/workflows/deploy.yml`); também dá para rodar à mão em *Actions → Deploy → Run workflow*,
com o modo simulação (`--check`).

## Como funciona

1. O CI compila o binário (`mise run build`, linux/amd64, sem CGO), com a SPA embutida e o modo
   mock desligado. A VPS tem 1 GB e não compila nada.
2. O Ansible (`deploy/ansible/deploy.yml`) envia o binário para `/opt/artifact-docs`, grava o
   `.env` (modo 600) a partir das secrets e sobe o container `artifact-docs` na rede `proxy`.
3. O container usa a imagem `distroless/static:nonroot`, sem shell, com o disco só de leitura e
   um teto de 160 MB (o Go com `GOMEMLIMIT=64MiB`). As migrations rodam ao subir.
4. O playbook espera o `/api/health` responder, grava `/opt/caddy/conf.d/artifact-docs.caddy`,
   valida e recarrega o Caddy.

Docker, a rede `proxy` e o Caddy são do my-config: rode a etapa `base` e a `caddy` dele antes do
primeiro deploy.

## Secrets (Environment `oracle` deste repositório)

| Secret | De onde vem |
|---|---|
| `SSH_PRIVATE_KEY` | a mesma do Environment `oracle` do my-config |
| `SSH_PASSPHRASE` | a mesma do my-config (vazia se a chave não tiver senha) |
| `ANSIBLE_BECOME_PASSWORD` | a mesma do my-config (opcional) |
| `ANSIBLE_HOSTS_INI` | a mesma do my-config — o playbook usa o grupo `[servers]` |
| `DATABASE_URL` | Neon → Connect → endpoint **pooled** |
| `NEON_AUTH_URL` | Neon → Auth → URL do endpoint (também vai no bundle; é pública) |
| `NEON_AUTH_JWKS_URL` | Neon → Auth → JWKS URL |
| `AUTH_ALLOWED_EMAILS` | quem pode entrar, separados por vírgula |
| `R2_ENDPOINT`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET` | Cloudflare → R2 → Manage API tokens |
| `ARTIFACT_LINK_SECRET` | `openssl rand -hex 32` |

O endereço público é `hub.vinicius-gpl.com`. Para trocar, crie a **variável** (não secret)
`APP_HOST` no repositório.

## Antes do primeiro deploy

- DNS: registro `A` de `hub` para o IP da Oracle na Cloudflare, em modo "DNS only", como os outros.
- Neon Auth: adicione `https://hub.vinicius-gpl.com` aos domínios confiáveis.
- Primeiro acesso: entre e crie a conta, e depois desligue o cadastro no Neon Auth (`docs/auth.md`).

## Localmente

```bash
mise run build   # com CGO_ENABLED=0 GOOS=linux GOARCH=amd64 e VITE_NEON_AUTH_URL no ambiente
cd deploy/ansible
ansible-playbook -i hosts.ini deploy.yml \
  -e binary_src=../../artifact-docs/bin/artifact-docs -e @vars.json
```

`vars.json` traz `app_host` e `app_env` (o mapa das variáveis acima). Nunca versione esse arquivo.
