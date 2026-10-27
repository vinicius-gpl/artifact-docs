-- O registro da versão publicada. A chave (artifact_id, version) é o que barra duas
-- publicações simultâneas de pegarem o mesmo número.
-- $1: artifact_id · $2: version · $3: size_bytes · $4: has_thumb
INSERT INTO artifact_versions (artifact_id, version, size_bytes, has_thumb)
VALUES ($1::UUID, $2, $3, $4);
