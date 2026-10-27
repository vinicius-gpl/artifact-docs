-- Os apps do hub em ordem alfabética, sem diferença de maiúscula. A tela os agrupa por tipo.
SELECT
  app.id::TEXT,
  app.type_id::TEXT,
  app.name,
  app.url,
  app.status,
  app.checked_at
FROM apps AS app
ORDER BY lower(app.name);
