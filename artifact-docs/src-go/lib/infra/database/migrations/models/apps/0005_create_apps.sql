-- Os apps do hub (briefing, módulo 4). O `status` é escrito pela goroutine que pinga cada
-- endereço; os valores são o appLinkStatusSchema de svelte/src/lib/contracts/app-link.schema.ts.

CREATE TABLE apps (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  type_id    UUID        NOT NULL,
  name       TEXT        NOT NULL,
  url        TEXT        NOT NULL,
  -- A chave da logo no R2 (apps/<id>/logo.<ext>). Sem ela, a tela usa o favicon do endereço.
  logo_key   TEXT,
  status     TEXT        NOT NULL DEFAULT 'unknown',
  checked_at TIMESTAMPTZ,

  CONSTRAINT apps_type_fkey    FOREIGN KEY (type_id) REFERENCES app_types (id) ON DELETE CASCADE,
  CONSTRAINT apps_name_filled  CHECK (name <> ''),
  CONSTRAINT apps_url_http     CHECK (url ~ '^https?://'),
  CONSTRAINT apps_status_known CHECK (status IN ('online', 'offline', 'unknown'))
);

CREATE INDEX apps_type_idx ON apps (type_id);
