-- Edita o app. Endereço novo volta para "sem verificação" e sem ícone: a situação medida e o
-- ícone achado eram do antigo.
-- $1: id · $2: type_id · $3: name · $4: url
--
-- Do lado direito do SET, `url` ainda é o valor de ANTES da mudança.
UPDATE apps
SET
  type_id    = $2::UUID,
  name       = $3,
  url        = $4,
  icon_url   = CASE WHEN url = $4 THEN icon_url ELSE NULL END,
  status     = CASE WHEN url = $4 THEN status ELSE 'unknown' END,
  checked_at = CASE WHEN url = $4 THEN checked_at ELSE NULL END
WHERE id = $1::UUID
RETURNING
  id::TEXT,
  type_id::TEXT,
  name,
  url,
  icon_url,
  status,
  checked_at;
