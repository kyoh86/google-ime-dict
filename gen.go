package gimedic

//go:generate go run -tags man ./cmd/gimedic man

// See: mise.toml
//go:generate protoc --proto_path=proto --go_out=. --go_opt=paths=source_relative user_dictionary_storage.proto
