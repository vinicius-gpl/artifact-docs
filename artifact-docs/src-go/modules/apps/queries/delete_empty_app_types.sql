-- O tipo que ficou sem nenhum app some: um título vazio no hub parece erro de carga.
DELETE FROM app_types AS app_type
WHERE NOT EXISTS (
  SELECT 1
  FROM apps AS app
  WHERE app.type_id = app_type.id
);
