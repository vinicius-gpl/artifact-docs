-- O que o monitor confere: todo app, com o endereço e o ícone que já tem (vazio = procurar).
SELECT
  app.id::TEXT,
  app.url,
  coalesce(app.icon_url, '')
FROM apps AS app;
