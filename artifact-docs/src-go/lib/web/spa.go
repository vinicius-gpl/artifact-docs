package web

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// notBuilt é o que aparece quando o binário foi compilado sem a SPA (ex.: `go run .` antes
// do primeiro `mise run build:web`). Em desenvolvimento a tela vem do Vite, na :5010.
const notBuilt = "A tela ainda não foi compilada. Em desenvolvimento, use a do Vite " +
	"(mise run dev → http://localhost:5010); para o binário, rode mise run build."

// Handler serve a SPA embutida.
//
// Arquivo que existe (JS, CSS, favicon) é servido como está. Qualquer outro caminho recebe o
// index.html: o SvelteKit está em modo SPA (adapter-static com fallback), e é o roteador dele,
// no navegador, que decide se /artifacts/briefing existe.
func Handler(dist fs.FS) http.Handler {
	files := http.FileServerFS(dist)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		name := strings.TrimPrefix(path.Clean(request.URL.Path), "/")

		if name != "" && isFile(dist, name) {
			files.ServeHTTP(writer, request)
			return
		}

		index, err := fs.ReadFile(dist, "index.html")
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(writer, notBuilt, http.StatusServiceUnavailable)
			return
		}
		if err != nil {
			http.Error(writer, "Não consegui abrir a tela.", http.StatusInternalServerError)
			return
		}

		// O index.html nunca vai para cache: é ele que aponta para os arquivos com hash de
		// cada build. Em cache, um deploy novo continuaria abrindo a tela antiga.
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write(index)
	})
}

func isFile(dist fs.FS, name string) bool {
	info, err := fs.Stat(dist, name)

	return err == nil && !info.IsDir()
}
