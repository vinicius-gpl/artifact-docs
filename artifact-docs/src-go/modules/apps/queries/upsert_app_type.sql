-- O tipo pelo nome: devolve o que existe ou cria um novo no fim da fila (`position`).
-- $1: name
--
-- O DO UPDATE não muda nada de verdade: está ali porque o DO NOTHING não devolve a linha que
-- já existia, e o cadastro precisa do id dela.
INSERT INTO app_types (name, position)
VALUES (
  $1,
  (SELECT coalesce(max(app_type.position) + 1, 0) FROM app_types AS app_type)
)
ON CONFLICT ((lower(name))) DO UPDATE
SET name = app_types.name
RETURNING id::TEXT;
