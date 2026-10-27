-- O histórico de versões de um artefato, da mais nova à mais antiga.
-- $1: artifact_id
SELECT
  version.version,
  version.size_bytes,
  version.has_thumb,
  version.created_at
FROM artifact_versions AS version
WHERE version.artifact_id = $1::UUID
ORDER BY version.version DESC;
