-- A versão nova de um artefato que já existe: troca o ponteiro e o que veio junto.
-- $1: id · $2: title (NULL = mantém) · $3: a descrição veio? · $4: description ('' limpa)
-- $5: kind (NULL = mantém) · $6: tags (NULL = mantém) · $7: format · $8: content_text
-- $9: current_version · $10: status (NULL = mantém) · $11: is_private (NULL = mantém)
UPDATE artifacts
SET
  title           = coalesce($2::TEXT, title),
  description     = CASE WHEN $3::BOOLEAN THEN nullif($4::TEXT, '') ELSE description END,
  kind            = coalesce($5::TEXT, kind),
  tags            = coalesce($6::TEXT[], tags),
  format          = $7,
  content_text    = $8,
  current_version = $9,
  status          = coalesce($10::TEXT, status),
  is_private      = coalesce($11::BOOLEAN, is_private),
  updated_at      = now()
WHERE id = $1::UUID;
