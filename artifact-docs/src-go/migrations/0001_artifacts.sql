-- Artefatos e as versões de cada um (briefing, módulos 1 a 3).
--
-- Documento, planejamento e ideia são a mesma tabela com um `kind` diferente: compartilham
-- galeria, busca, preview e versões. Os valores de `kind`, `format` e `status` são os mesmos
-- enums de svelte/src/lib/contracts/artifact.schema.ts — mudou lá, muda aqui.
--
-- Não há tabela de usuário: quem entra é o Neon Auth, e `owner_id` guarda o id de lá. Ele
-- existe desde já pensando na ida para a empresa (multiusuário).

create table artifacts (
  id              uuid primary key default gen_random_uuid(),
  owner_id        text not null,
  label           text not null check (label ~ '^[a-z0-9]+(-[a-z0-9]+)*$' and length(label) <= 80),
  title           text not null check (title <> ''),
  description     text,
  kind            text not null check (kind in ('document', 'plan', 'idea')),
  format          text not null check (format in ('html', 'md', 'pdf')),
  status          text check (status in ('draft', 'active', 'done', 'archived')),
  -- O segredo do link /a/<token>/. Aleatório, e não derivado da label: quem souber a label
  -- não consegue calcular o link (briefing, "Link do artefato"). Dois UUID v4 sem hífen são
  -- 244 bits aleatórios, sem depender da extensão pgcrypto.
  link_token      text not null unique
                  default replace(gen_random_uuid()::text || gen_random_uuid()::text, '-', ''),
  is_private      boolean not null default false,
  is_pinned       boolean not null default false,
  -- O ponteiro da versão servida: o rollback é só trocar este número.
  current_version integer not null default 1 check (current_version > 0),
  tags            text[] not null default '{}',
  -- O texto extraído no upload, só para a busca; nunca vai para a tela.
  content_text    text not null default '',
  search          tsvector not null default ''::tsvector,
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now(),
  last_viewed_at  timestamptz,
  unique (owner_id, label)
);

-- A busca em português (briefing, "Neon: cuidados"). Título pesa mais que descrição e tags,
-- que pesam mais que o corpo.
create function artifacts_search_refresh() returns trigger language plpgsql as $$
begin
  new.search :=
    setweight(to_tsvector('portuguese', new.title), 'A') ||
    setweight(to_tsvector('portuguese', coalesce(new.description, '') || ' ' || array_to_string(new.tags, ' ')), 'B') ||
    setweight(to_tsvector('portuguese', new.content_text), 'C');
  return new;
end
$$;

create trigger artifacts_search_refresh
  before insert or update of title, description, tags, content_text on artifacts
  for each row execute function artifacts_search_refresh();

create index artifacts_search_idx on artifacts using gin (search);
create index artifacts_owner_updated_idx on artifacts (owner_id, updated_at desc);

create table artifact_versions (
  artifact_id uuid not null references artifacts (id) on delete cascade,
  version     integer not null check (version > 0),
  size_bytes  bigint not null check (size_bytes >= 0),
  has_thumb   boolean not null default false,
  created_at  timestamptz not null default now(),
  primary key (artifact_id, version)
);

-- Hub de apps (briefing, módulo 4). Os tipos são criados pela pessoa; os apps aparecem
-- agrupados por eles, na ordem de `position`.
create table app_types (
  id       uuid primary key default gen_random_uuid(),
  name     text not null check (name <> ''),
  position integer not null default 0 check (position >= 0)
);

create table apps (
  id         uuid primary key default gen_random_uuid(),
  type_id    uuid not null references app_types (id) on delete cascade,
  name       text not null check (name <> ''),
  url        text not null check (url ~ '^https?://'),
  -- A chave da logo no R2 (apps/<id>/logo.<ext>). Sem ela, a tela usa o favicon do endereço.
  logo_key   text,
  status     text not null default 'unknown' check (status in ('online', 'offline', 'unknown')),
  checked_at timestamptz
);

create index apps_type_idx on apps (type_id);
