-- O ícone do app, descoberto pelo monitor na página do próprio endereço (o <link rel="icon">,
-- ou o /favicon.ico). Fica o endereço absoluto: a tela carrega direto do site, sem passar pelo
-- R2 nem pela VPS.
--
-- Separado do `logo_key`: aquele é a logo que a pessoa escolher; este é o que o monitor achou, e
-- pode ser reescrito a cada conferência sem apagar escolha de ninguém.
ALTER TABLE apps ADD COLUMN icon_url TEXT;
