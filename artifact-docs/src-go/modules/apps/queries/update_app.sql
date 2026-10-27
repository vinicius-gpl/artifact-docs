-- Edita o app. Endereço novo volta para "sem verificação": a situação medida era do antigo.
-- $1: id · $2: type_id · $3: name · $4: url
--
-- Do lado direito do SET, `url` ainda é o valor de ANTES da mudança.
UPDATE apps
SET
  type_id    = $2::UUID,
  name       = $3,
  url        = $4,
  status     = CASE WHEN url = $4 THEN status ELSE 'unknown' END,
  checked_at = CASE WHEN url = $4 THEN checked_at ELSE NULL END
WHERE id = $1::UUID
RETURNING
  id::TEXT,
  type_id::TEXT,
  name,
  url,
  status,
  checked_at;
