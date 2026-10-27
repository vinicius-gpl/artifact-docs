package apps

import (
	"strings"
	"testing"
)

// feliz
func TestNormalizeTrimsAndAcceptsAWebApp(test *testing.T) {
	input, problem := normalize(Input{Name: " Neon ", URL: " https://console.neon.tech ", TypeName: " Ferramentas "})

	if problem != "" || input.Name != "Neon" || input.URL != "https://console.neon.tech" || input.TypeName != "Ferramentas" {
		test.Fatalf("input = %+v, problem = %q", input, problem)
	}
}

// triste
func TestNormalizeRefusesWhatIsNotAnAppAddress(test *testing.T) {
	for _, address := range []string{"neon.tech", "javascript:alert(1)", "https://", "https://" + strings.Repeat("a", 2100)} {
		if _, problem := normalize(Input{Name: "X", URL: address, TypeName: "T"}); problem == "" {
			test.Errorf("%q: expected a problem", address)
		}
	}
}
