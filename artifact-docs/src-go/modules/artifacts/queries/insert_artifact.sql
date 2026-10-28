-- O artefato novo, já apontando para a versão 1. O id vem do Go: os arquivos sobem para o
-- R2 (artifacts/<id>/v1/) ANTES desta linha existir — a publicação só aparece quando tudo
-- já está lá.
-- $1: id · $2: owner_id · $3: label · $4: title · $5: description · $6: kind · $7: format
-- $8: tags · $9: content_text · $10: current_version · $11: status · $12: is_private
INSERT INTO artifacts (
  id,
  owner_id,
  label,
  title,
  description,
  kind,
  format,
  tags,
  content_text,
  current_version,
  status,
  is_private
)
VALUES (
  $1::UUID, $2, $3, $4, nullif($5::TEXT, ''), $6, $7, coalesce($8::TEXT[], '{}'), $9, $10,
  $11::TEXT, coalesce($12::BOOLEAN, TRUE)
);
