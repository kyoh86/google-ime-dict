package syncer

import (
	"os"

	"github.com/kyoh86/gimedic"
	"google.golang.org/protobuf/proto"
)

func LoadStorage(path string) (*gimedic.UserDictionaryStorage, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var storage gimedic.UserDictionaryStorage
	if err := proto.Unmarshal(raw, &storage); err != nil {
		return nil, err
	}
	return &storage, nil
}

func WriteStorage(path string, storage *gimedic.UserDictionaryStorage) error {
	raw, err := proto.Marshal(storage)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
