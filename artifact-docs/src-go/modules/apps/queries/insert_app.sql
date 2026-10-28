-- O app novo. Nasce "sem verificação" e sem ícone: quem mede e acha o ícone é o monitor, logo
-- em seguida.
-- $1: type_id · $2: name · $3: url
INSERT INTO apps (type_id, name, url)
VALUES ($1::UUID, $2, $3)
RETURNING
  id::TEXT,
  type_id::TEXT,
  name,
  url,
  icon_url,
  status,
  checked_at;
