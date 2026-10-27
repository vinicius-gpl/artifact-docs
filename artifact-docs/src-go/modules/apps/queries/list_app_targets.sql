-- O que o monitor confere: todo app, com o endereço.
SELECT
  app.id::TEXT,
  app.url
FROM apps AS app;
