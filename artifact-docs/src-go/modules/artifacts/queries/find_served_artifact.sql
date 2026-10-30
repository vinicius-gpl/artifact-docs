-- O artefato por trás de um link /a/<token>/: qual versão servir, se ele é privado e o que
-- vai no cartão do link (título, descrição, tipo e se a versão servida tem miniatura).
-- $1: link_token
SELECT
  summary.id::TEXT,
  summary.current_version,
  summary.format,
  summary.is_private,
  summary.title,
  summary.description,
  summary.kind,
  summary.has_thumb
FROM artifact_summary_view AS summary
WHERE summary.link_token = $1::TEXT;
