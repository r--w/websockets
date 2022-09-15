## NATS

[Installation](https://docs.nats.io/running-a-nats-service/introduction/installation)

Running: `nats-server -m 8222`

Monitoring: [http://localhost:8222/](http://localhost:8222/)

## Protobuf [docs](https://developers.google.com/protocol-buffers/docs/gotutorial)

1. `brew install protobuf   `
2. `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
3. `protoc --go_out=. ./pb/ticker.proto`

### Nats + protobuf encoder
1. [example](https://github.com/nats-io/nats.go/blob/main/test/protobuf_test.go)

