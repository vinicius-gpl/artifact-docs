-- O ícone que o monitor achou na página do app.
-- $1: id · $2: icon_url
UPDATE apps
SET icon_url = $2
WHERE id = $1::UUID;
