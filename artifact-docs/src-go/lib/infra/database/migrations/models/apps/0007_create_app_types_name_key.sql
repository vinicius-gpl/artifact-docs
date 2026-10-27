-- Um nome de tipo, um tipo só — sem diferença de maiúscula ("Na VPS" e "na vps" são o mesmo).
--
-- É o que deixa o cadastro de app mandar o tipo pelo NOME: o tipo que não existe nasce na
-- hora, e o que existe é reaproveitado (INSERT … ON CONFLICT), sem tela para gerenciar tipos.
CREATE UNIQUE INDEX app_types_name_key ON app_types (lower(name));
