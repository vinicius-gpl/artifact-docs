-- Cada upload é uma versão nova (briefing, módulo 1). Os arquivos moram no R2 em
-- artifacts/<artifact_id>/v<version>/; aqui fica só o que a galeria mostra sobre eles.

CREATE TABLE artifact_versions (
  artifact_id UUID        NOT NULL,
  version     INTEGER     NOT NULL,
  size_bytes  BIGINT      NOT NULL,
  has_thumb   BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT artifact_versions_pkey             PRIMARY KEY (artifact_id, version),
  CONSTRAINT artifact_versions_artifact_fkey    FOREIGN KEY (artifact_id) REFERENCES artifacts (id) ON DELETE CASCADE,
  CONSTRAINT artifact_versions_version_positive CHECK (version > 0),
  CONSTRAINT artifact_versions_size_positive    CHECK (size_bytes >= 0)
);
