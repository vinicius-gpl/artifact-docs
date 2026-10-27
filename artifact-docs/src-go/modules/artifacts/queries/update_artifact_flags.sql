-- O que muda sem subir versão nova: fixar, privado e situação (artifactPatchSchema).
-- $1: owner_id · $2: label · $3: is_pinned (NULL = mantém) · $4: is_private (NULL = mantém)
-- $5: o status veio no corpo? · $6: o status novo (NULL limpa)
--
-- O status precisa do $5 à parte porque, nele, NULL é um valor ("sem situação"), e não
-- "mantém". Não mexe em updated_at: fixar um card não é editar o artefato, e o "Editado há X"
-- da galeria precisa continuar dizendo quando o conteúdo mudou.
UPDATE artifacts
SET
  is_pinned  = coalesce($3::BOOLEAN, is_pinned),
  is_private = coalesce($4::BOOLEAN, is_private),
  status     = CASE WHEN $5::BOOLEAN THEN $6::TEXT ELSE status END
WHERE owner_id = $1
  AND label    = $2;
