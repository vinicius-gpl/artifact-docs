-- Os tipos do hub, na ordem que a pessoa escolheu; empate cai no nome.
SELECT
  app_type.id::TEXT,
  app_type.name,
  app_type.position
FROM app_types AS app_type
ORDER BY app_type.position, app_type.name;
