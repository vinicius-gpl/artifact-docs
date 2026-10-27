-- O rollback: o ponteiro volta para uma versão que já existe (briefing, módulo 1). Os
-- arquivos de todas as versões continuam no R2, então nada é copiado.
-- $1: owner_id · $2: label · $3: version
UPDATE artifacts AS artifact
SET
  current_version = $3,
  updated_at      = now()
WHERE artifact.owner_id = $1
  AND artifact.label    = $2
  AND EXISTS (
    SELECT 1
    FROM artifact_versions AS version
    WHERE version.artifact_id = artifact.id
      AND version.version     = $3
  )
RETURNING artifact.link_token;
