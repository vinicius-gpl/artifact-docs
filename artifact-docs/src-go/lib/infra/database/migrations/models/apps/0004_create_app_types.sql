-- Os tipos do hub de apps (briefing, módulo 4). São criados pela pessoa, e os apps aparecem
-- agrupados por eles, na ordem de `position`.

CREATE TABLE app_types (
  id       UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  name     TEXT    NOT NULL,
  position INTEGER NOT NULL DEFAULT 0,

  CONSTRAINT app_types_name_filled       CHECK (name <> ''),
  CONSTRAINT app_types_position_positive CHECK (position >= 0)
);
