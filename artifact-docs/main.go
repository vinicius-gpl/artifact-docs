// Command artifact-docs sobe o hub: a API em /api e a SPA embutida em todo o resto.
//
// Um binário só, sem Node em produção: a VPS tem 1 GB de RAM, e cada processo a menos é
// memória a mais para o resto. Todo o código Go mora em src-go/ (app/ compõe, modules/ tem
// as funcionalidades, lib/ a infraestrutura); aqui só se cria o log e se chama o servidor.
package main

import (
	"log/slog"
	"os"

	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/app/server"
	"github.com/vinicius-gpl/artifact-docs/artifact-docs/src-go/lib/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := server.Run(logger, web.Dist()); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
