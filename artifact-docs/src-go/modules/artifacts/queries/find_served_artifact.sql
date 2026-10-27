-- O artefato por trás de um link /a/<token>/: qual versão servir e se ele é privado.
-- $1: link_token
SELECT
  artifact.id::TEXT,
  artifact.current_version,
  artifact.format,
  artifact.is_private
FROM artifacts AS artifact
WHERE artifact.link_token = $1;
