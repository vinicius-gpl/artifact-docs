-- O que muda sem subir versão nova (artifactPatchSchema): fixar, privado, situação, título,
-- descrição e tags.
-- $1: owner_id · $2: label · $3: is_pinned (NULL = mantém) · $4: is_private (NULL = mantém)
-- $5: o status veio no corpo? · $6: o status novo (NULL limpa)
-- $7: title (NULL = mantém) · $8: a descrição veio? · $9: description ('' ou NULL limpa)
-- $10: tags (NULL = mantém; vazia limpa)
--
-- Status e descrição precisam do "veio?" à parte porque, neles, NULL é um valor ("sem
-- situação", "sem descrição"), e não "mantém". Não mexe em updated_at: o "Editado há X" da
-- galeria diz quando o CONTEÚDO (a versão) mudou, não o cartão. A busca acompanha sozinha: o
-- gatilho artifacts_search_refresh recalcula quando título, descrição ou tags mudam.
UPDATE artifacts
SET
  is_pinned   = coalesce($3::BOOLEAN, is_pinned),
  is_private  = coalesce($4::BOOLEAN, is_private),
  status      = CASE WHEN $5::BOOLEAN THEN $6::TEXT ELSE status END,
  title       = coalesce($7::TEXT, title),
  description = CASE WHEN $8::BOOLEAN THEN nullif($9::TEXT, '') ELSE description END,
  tags        = coalesce($10::TEXT[], tags)
WHERE owner_id = $1
  AND label    = $2;
