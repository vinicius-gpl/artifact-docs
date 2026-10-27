# Login — Neon Auth com e-mail e senha

O hub é de uma pessoa só. Entrar é **e-mail e senha no Neon Auth**, direto do navegador; o Go
nunca vê a senha, só confere o token que a Neon assina.

## Como funciona

1. **Primeiro acesso** (só enquanto não existe a conta): a tela de login mostra o botão
   "Primeiro acesso", e a conta nasce no Neon Auth (`signUp.email`) já logada.
2. **Entrar**: a tela manda e-mail e senha para o Neon Auth (`signIn.email`) e recebe a sessão.
3. **Cada chamada à API** leva `Authorization: Bearer <token>` (o JWT da Neon).
4. **O porteiro do Go** (middleware em toda rota `/api/*`, menos as públicas):
   - token ausente, inválido ou expirado → **401** → a tela manda para o login;
   - token válido, mas e-mail fora de `AUTH_ALLOWED_EMAILS` → **403** → login com
     "esta conta não tem acesso".

A regra fica no Go, não na tela: a tela só esconde; quem impede é o servidor.

## O que o Go precisa implementar

| Rota | Pública | O que faz |
|---|---|---|
| `GET /api/auth/setup` | sim | `{ "needsSetup": true }` enquanto não há usuário no Neon Auth com e-mail de `AUTH_ALLOWED_EMAILS` (consulta em `neon_auth.user`) |
| `GET /api/auth/me` | não | `{ id, email, name, image }` de quem está no token |
| `GET /api/auth/config` | sim | `{ "neonAuthUrl": "…" }` — onde o login acontece, quando a URL não veio no build da tela |
| `GET /api/health` | sim | já existe |

O porteiro valida a assinatura do JWT com as chaves públicas do Neon Auth (JWKS, buscadas uma
vez e guardadas em memória, renovadas quando aparece um `kid` desconhecido) e confere `exp`.

Variáveis do Go (`src-go/.env`):

```
NEON_AUTH_URL=          # Auth → URL do endpoint (entregue em /api/auth/config)
NEON_AUTH_JWKS_URL=     # no painel da Neon: Auth → JWKS URL (vazia = login desligado, 503)
AUTH_ALLOWED_EMAILS=    # o seu e-mail; vários separados por vírgula
```

Variável da tela (`svelte/.env`): `VITE_NEON_AUTH_URL` (Auth → URL do endpoint). É pública:
só diz ONDE o login acontece.

## Segurança

- **Cadastro aberto na Neon não abre o hub**: quem criar conta sem estar na lista recebe 403
  em tudo. Mesmo assim, **depois do primeiro acesso, desligue novos cadastros** no painel do
  Neon Auth.
- **Mesma mensagem** para e-mail inexistente e senha errada ("E-mail ou senha incorretos.").
- **Sair limpa o cache** de dados da tela antes de ir para o login.
- **Artefato privado** (`/a/<token>/`): página aberta direto não manda o cabeçalho
  `Authorization`. Quando os privados existirem, a API autenticada devolve um link assinado e
  de vida curta para eles. Os públicos abrem só com o link.
- **Skills** que publicam pela API usam um token de API próprio (com hash no banco), não a
  sessão de login.

## No e2e (APP_ENV=test)

O e2e roda a tela contra o **Go de verdade**, mas sem login de verdade: com `APP_ENV=test`, o
porteiro deixa toda requisição entrar como o **usuário de teste** — o mesmo do modo mock da
tela (`user-mock` / `voce@exemplo.com`, `auth.TestUser` em `src-go/modules/auth/guard.go`).
Nenhum token é pedido e o Neon Auth não é consultado; o `/api/auth/setup` responde
`needsSetup: false`.

Para isso nunca vazar para produção, o Go **recusa subir** com `APP_ENV=test` e
`NEON_AUTH_JWKS_URL` preenchidas juntas, e avisa no log ao subir em modo de teste.

## Em localhost (modo mock)

Com `VITE_API_MOCKS=true` (padrão do `mise run dev`), nada disso sai da máquina: o
`src/lib/api/mock-auth.ts` imita o Neon Auth e o porteiro.

- O formulário já vem com o **usuário de teste** (`voce@exemplo.com` / `artifact-docs`).
- O **primeiro acesso** aparece até você criar uma conta pela tela; depois, ela também entra.
- Sessão e conta ficam no `localStorage` do navegador. Para recomeçar, limpe os dados do site.
