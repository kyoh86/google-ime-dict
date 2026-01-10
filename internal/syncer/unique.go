package syncer

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/kyoh86/gimedic"
)

func UniqueDictionaryID(storage *gimedic.UserDictionaryStorage) uint64 {
	used := map[uint64]struct{}{}
	for _, d := range storage.GetDictionaries() {
		used[d.GetId()] = struct{}{}
	}
	for {
		id := randomUint64()
		if id == 0 {
			continue
		}
		if _, ok := used[id]; !ok {
			return id
		}
	}
}

func randomUint64() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0
	}
	return binary.LittleEndian.Uint64(b[:])
}
