-- O artefato: documento, planejamento ou ideia (briefing, módulos 1 a 3).
--
-- Os três são a mesma tabela com um `kind` diferente: compartilham galeria, busca, preview e
-- versões. Os valores de `kind`, `format` e `status` são os enums de
-- svelte/src/lib/contracts/artifact.schema.ts — mudou lá, muda aqui.
--
-- Não há tabela de usuário: quem entra é o Neon Auth, e `owner_id` guarda o id de lá. Ele
-- existe desde já pensando na ida para a empresa (multiusuário).

CREATE TABLE artifacts (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id        TEXT        NOT NULL,
  label           TEXT        NOT NULL,
  title           TEXT        NOT NULL,
  description     TEXT,
  kind            TEXT        NOT NULL,
  format          TEXT        NOT NULL,
  status          TEXT,
  -- O segredo do link /a/<token>/. Aleatório, e não derivado da label: quem souber a label
  -- não consegue calcular o link. Dois UUID v4 sem hífen são 244 bits aleatórios, sem
  -- depender da extensão pgcrypto.
  link_token      TEXT        NOT NULL DEFAULT replace(gen_random_uuid()::TEXT || gen_random_uuid()::TEXT, '-', ''),
  is_private      BOOLEAN     NOT NULL DEFAULT FALSE,
  is_pinned       BOOLEAN     NOT NULL DEFAULT FALSE,
  -- O ponteiro da versão servida: o rollback é só trocar este número.
  current_version INTEGER     NOT NULL DEFAULT 1,
  tags            TEXT[]      NOT NULL DEFAULT '{}',
  -- O texto extraído no upload, só para a busca; nunca vai para a tela.
  content_text    TEXT        NOT NULL DEFAULT '',
  search          TSVECTOR    NOT NULL DEFAULT ''::TSVECTOR,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_viewed_at  TIMESTAMPTZ,

  CONSTRAINT artifacts_label_slug       CHECK (label ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(label) <= 80),
  CONSTRAINT artifacts_title_filled     CHECK (title <> ''),
  CONSTRAINT artifacts_kind_known       CHECK (kind IN ('document', 'plan', 'idea')),
  CONSTRAINT artifacts_format_known     CHECK (format IN ('html', 'md', 'pdf')),
  CONSTRAINT artifacts_status_known     CHECK (status IN ('draft', 'active', 'done', 'archived')),
  CONSTRAINT artifacts_version_positive CHECK (current_version > 0),
  CONSTRAINT artifacts_link_token_key   UNIQUE (link_token),
  CONSTRAINT artifacts_owner_label_key  UNIQUE (owner_id, label)
);

-- A galeria lista os artefatos do dono, do editado mais recente ao mais antigo.
CREATE INDEX artifacts_owner_updated_idx ON artifacts (owner_id, updated_at DESC);
