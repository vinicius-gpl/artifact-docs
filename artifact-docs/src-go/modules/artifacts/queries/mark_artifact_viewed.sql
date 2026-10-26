-- Abrir o artefato conta como acesso: é o que ordena "Recentes" na galeria.
-- $1: owner_id · $2: label
UPDATE artifacts
SET last_viewed_at = now()
WHERE owner_id = $1
  AND label    = $2;
