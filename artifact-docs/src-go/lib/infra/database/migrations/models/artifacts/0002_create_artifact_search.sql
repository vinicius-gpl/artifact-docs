-- A busca em português (briefing, "Neon: cuidados"): tsvector no dicionário `portuguese` com
-- índice GIN.
--
-- É gatilho, e não coluna gerada, porque `array_to_string` não é IMMUTABLE e o Postgres
-- recusa coluna gerada que dependa dela. O peso ordena a relevância: título (A) vale mais que
-- descrição e tags (B), que valem mais que o corpo do documento (C).

CREATE FUNCTION artifacts_search_refresh() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
  NEW.search :=
       setweight(to_tsvector('portuguese', NEW.title), 'A')
    || setweight(to_tsvector('portuguese', coalesce(NEW.description, '') || ' ' || array_to_string(NEW.tags, ' ')), 'B')
    || setweight(to_tsvector('portuguese', NEW.content_text), 'C');

  RETURN NEW;
END
$$;

CREATE TRIGGER artifacts_search_refresh
  BEFORE INSERT OR UPDATE OF title, description, tags, content_text ON artifacts
  FOR EACH ROW
  EXECUTE FUNCTION artifacts_search_refresh();

CREATE INDEX artifacts_search_idx ON artifacts USING GIN (search);
