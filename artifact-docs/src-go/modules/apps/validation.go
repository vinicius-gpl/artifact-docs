package apps

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Os limites do appLinkInputSchema (svelte/src/lib/contracts/app-link.schema.ts) — mudou lá,
// muda aqui.
const (
	maxName     = 80
	maxTypeName = 40
	maxURL      = 2048
)

// idPattern é o formato do id (UUID). Id fora dele nem vai ao banco: é 404 direto, em vez de
// um erro de conversão do Postgres virando 500.
var idPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// normalize limpa os espaços e confere o cadastro. A mensagem é a que a tela mostra.
func normalize(input Input) (Input, string) {
	input = Input{
		Name:     strings.TrimSpace(input.Name),
		URL:      strings.TrimSpace(input.URL),
		TypeName: strings.TrimSpace(input.TypeName),
	}

	if input.Name == "" || utf8.RuneCountInString(input.Name) > maxName {
		return input, "Informe o nome do app (até 80 caracteres)."
	}
	if !isWebAddress(input.URL) {
		return input, "Informe o endereço completo, começando com http:// ou https://."
	}
	if input.TypeName == "" || utf8.RuneCountInString(input.TypeName) > maxTypeName {
		return input, "Informe o tipo do app (até 40 caracteres)."
	}

	return input, ""
}

func isWebAddress(raw string) bool {
	if len(raw) > maxURL {
		return false
	}
	address, err := url.Parse(raw)

	return err == nil && (address.Scheme == "http" || address.Scheme == "https") && address.Host != ""
}
