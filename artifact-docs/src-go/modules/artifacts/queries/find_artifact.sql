-- Um artefato do dono, pela label. Mesmas colunas, na mesma ordem, de list_artifacts.sql:
-- as duas passam pelo mesmo scan.
-- $1: owner_id · $2: label
SELECT
  summary.id::TEXT,
  summary.label,
  summary.title,
  summary.description,
  summary.kind,
  summary.format,
  summary.status,
  summary.link_token,
  summary.is_private,
  summary.is_pinned,
  summary.current_version,
  summary.size_bytes,
  summary.has_thumb,
  summary.tags,
  summary.created_at,
  summary.updated_at,
  summary.last_viewed_at
FROM artifact_summary_view AS summary
WHERE summary.owner_id = $1
  AND summary.label    = $2;
