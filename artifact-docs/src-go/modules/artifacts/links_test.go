package artifacts

import (
	"strings"
	"testing"
	"time"
)

func linkerAt(moment time.Time) *Linker {
	linker := NewLinker([]byte("test-secret"))
	linker.now = func() time.Time { return moment }

	return linker
}

var linkNow = time.Date(2026, 10, 26, 13, 0, 0, 0, time.UTC)

// feliz
func TestAPublicArtifactLinkIsJustTheToken(test *testing.T) {
	linker := linkerAt(linkNow)

	if url := linker.URL("abc", false); url != "/a/abc/" {
		test.Fatalf("url = %q", url)
	}
	if access := linker.Read("abc"); access.LinkToken != "abc" || access.Signed {
		test.Fatalf("access = %+v, want the token, unsigned", access)
	}
}

// feliz
func TestAPrivateLinkIsSignedAndStableInsideTheWindow(test *testing.T) {
	first := linkerAt(linkNow).URL("abc", true)
	later := linkerAt(linkNow.Add(time.Hour)).URL("abc", true)

	if first != later {
		test.Fatalf("links changed inside the window: %q then %q (the iframe would reload)", first, later)
	}

	access := linkerAt(linkNow).Read(strings.Trim(strings.TrimPrefix(first, "/a/"), "/"))
	if access.LinkToken != "abc" || !access.Signed {
		test.Fatalf("access = %+v, want a valid signature", access)
	}
}

// triste
func TestASignedLinkStopsWorkingWhenExpiredOrTampered(test *testing.T) {
	segment := strings.Trim(strings.TrimPrefix(linkerAt(linkNow).URL("abc", true), "/a/"), "/")

	if linkerAt(linkNow.Add(3 * linkWindow)).Read(segment).Signed {
		test.Error("an expired link must not be signed")
	}

	tampered := strings.Replace(segment, "abc", "xyz", 1)
	if linkerAt(linkNow).Read(tampered).Signed {
		test.Error("a link signed for another token must not be signed")
	}

	otherSecret := NewLinker([]byte("another-secret"))
	otherSecret.now = func() time.Time { return linkNow }
	if otherSecret.Read(segment).Signed {
		test.Error("a link signed with another secret must not be signed")
	}
}
