// Package web guarda a SPA compilada dentro do binário.
//
// `mise run build:web` compila o svelte/ e copia o resultado para dist/, ao lado deste
// arquivo. O go:embed só enxerga arquivos dentro do próprio pacote — por isso a cópia, em vez
// de apontar direto para svelte/dist.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist é a SPA compilada, com index.html na raiz. Antes do primeiro build ela só tem o
// .gitkeep, e o servidor responde que a tela ainda não foi compilada.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		// "dist" é um literal válido e está no embed: só falha se o código acima mudar.
		panic(err)
	}

	return sub
}
