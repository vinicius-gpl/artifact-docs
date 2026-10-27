-- O artefato como a galeria o vê: a linha de `artifacts` mais o tamanho e a miniatura da
-- versão SERVIDA (a de `current_version`).
--
-- Mora numa view para que toda consulta de artefato leia do mesmo lugar, em vez de repetir
-- o mesmo JOIN em cada uma. A view é expandida pelo Postgres na consulta: o filtro em
-- `search` continua usando o índice GIN da tabela.

CREATE VIEW artifact_summary_view AS
SELECT
  artifact.id,
  artifact.owner_id,
  artifact.label,
  artifact.title,
  artifact.description,
  artifact.kind,
  artifact.format,
  artifact.status,
  artifact.link_token,
  artifact.is_private,
  artifact.is_pinned,
  artifact.current_version,
  coalesce(served.size_bytes, 0)    AS size_bytes,
  coalesce(served.has_thumb, FALSE) AS has_thumb,
  artifact.tags,
  artifact.search,
  artifact.created_at,
  artifact.updated_at,
  artifact.last_viewed_at
FROM artifacts AS artifact
LEFT JOIN artifact_versions AS served
  ON  served.artifact_id = artifact.id
  AND served.version     = artifact.current_version;
