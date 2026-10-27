-- A galeria: os artefatos do dono, do editado mais recente ao mais antigo.
-- $1: owner_id · $2: kind ('' = todos) · $3: busca ('' = sem busca)
--
-- A busca usa websearch_to_tsquery: aceita o que a pessoa digita ("cnpj -abertura", frases
-- entre aspas) sem nunca dar erro de sintaxe.
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
  AND ($2::TEXT = '' OR summary.kind = $2::TEXT)
  AND ($3::TEXT = '' OR summary.search @@ websearch_to_tsquery('portuguese', $3::TEXT))
ORDER BY summary.updated_at DESC;
