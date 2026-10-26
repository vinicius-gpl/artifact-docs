-- Alguém da lista permitida (AUTH_ALLOWED_EMAILS) já criou a conta no Neon Auth?
-- $1: os e-mails permitidos, já em minúsculas.
--
-- "user" vai entre aspas: é palavra reservada do Postgres, e é o nome que o Neon Auth usa.
SELECT EXISTS (
  SELECT 1
  FROM neon_auth."user" AS account
  WHERE lower(account.email) = ANY ($1::TEXT[])
);
