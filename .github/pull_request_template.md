## O que muda

<!-- Uma ou duas frases, em português, sobre o que passa a dar para fazer no hub. -->

## Como conferir

<!-- O passo a passo para ver funcionando: qual tela abrir, o que clicar, o que deve aparecer. -->

## Definition of Done

- [ ] Título do PR no formato `[merge](local): Mensagem`
- [ ] Commits no formato `[tipo](local): Mensagem`
- [ ] Branch saiu de `develop` (ou de `main`, se for `hotfix/`)
- [ ] CI do PR verde (testes, lint, typecheck, design e build)
- [ ] Caminho feliz **e** caminho triste testados (Vitest e `go test`)
- [ ] Story para todo componente `acerola-*` novo ou alterado
- [ ] Nada editado dentro de `lib/components/ui/` nem `lib/hooks/ui/`
- [ ] `.env.example` atualizado se entrou variável nova (`artifact-docs/svelte/` ou `artifact-docs/src-go/`)
- [ ] Nenhuma chave, senha ou dado real no diff — inclusive em story e teste
- [ ] Idioma conferido: nada em inglês na tela, nada em português em log/erro interno/teste
