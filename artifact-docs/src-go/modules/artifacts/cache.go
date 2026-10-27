package artifacts

import (
	"sync"
	"time"
)

// Uma página publicada pede o index.html e, em seguida, dezenas de assets — todos com o mesmo
// token. Sem cache, cada asset seria uma consulta ao Neon (que fica na rede, e acorda devagar
// no plano free). O cache guarda "token → qual versão servir" por pouco tempo, e quem muda
// a versão (publicar, restaurar, apagar, mudar o privado) esquece o token na hora.

// servedTTL é quanto um token fica no cache. Curto: o que muda por fora (outro processo,
// alguém editando o banco à mão) aparece logo.
const servedTTL = 30 * time.Second

// maxServedEntries limita a memória. Passou, o cache recomeça vazio — mais simples que um
// LRU, e com um dono só quase nunca acontece.
const maxServedEntries = 512

type servedEntry struct {
	served  Served
	expires time.Time
}

type servedCache struct {
	mutex   sync.Mutex
	entries map[string]servedEntry
	now     func() time.Time
}

func newServedCache() *servedCache {
	return &servedCache{entries: map[string]servedEntry{}, now: time.Now}
}

func (cache *servedCache) get(linkToken string) (Served, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	entry, ok := cache.entries[linkToken]
	if !ok || cache.now().After(entry.expires) {
		return Served{}, false
	}

	return entry.served, true
}

func (cache *servedCache) put(linkToken string, served Served) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if len(cache.entries) >= maxServedEntries {
		cache.entries = map[string]servedEntry{}
	}
	cache.entries[linkToken] = servedEntry{served: served, expires: cache.now().Add(servedTTL)}
}

func (cache *servedCache) forget(linkToken string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	delete(cache.entries, linkToken)
}
