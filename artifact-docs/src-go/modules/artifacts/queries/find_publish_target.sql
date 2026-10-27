-- Onde a próxima versão vai cair: o artefato do dono com esta label (se existir) e a maior
-- versão já publicada. A próxima é essa + 1 — mesmo depois de um rollback, um número já
-- usado nunca é reaproveitado.
-- $1: owner_id · $2: label
SELECT
  artifact.id::TEXT,
  artifact.link_token,
  coalesce(max(version.version), 0) AS latest_version
FROM artifacts AS artifact
LEFT JOIN artifact_versions AS version
  ON version.artifact_id = artifact.id
WHERE artifact.owner_id = $1
  AND artifact.label    = $2
GROUP BY artifact.id;
