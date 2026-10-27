-- Apaga o artefato e, em cascata, as versões. Devolve o que o Go precisa para apagar os
-- arquivos no R2 e esquecer o link no cache.
-- $1: owner_id · $2: label
DELETE FROM artifacts AS artifact
WHERE artifact.owner_id = $1
  AND artifact.label    = $2
RETURNING artifact.id::TEXT, artifact.link_token;
