.PHONY: test

example:
	go run main.go -script ./examples/script.sql --driver postgres --conn postgres://root@localhost:26257/sandbox?sslmode=disable

debug:
	go run main.go -debug -script ./examples/sandbox.sql --driver postgres --conn postgres://root@localhost:26257/sandbox?sslmode=disable

test:
	go test ./... -v ;\
	go test ./... -cover

bench:
	go test ./... -bench=.

cover:
	go test ./... -coverprofile=coverage.out -coverpkg=\
	github.com/codingconcepts/datagen/internal/pkg/parse,\
	github.com/codingconcepts/datagen/internal/pkg/random,\
	github.com/codingconcepts/datagen/internal/pkg/runner;\
	go tool cover -html=coverage.out