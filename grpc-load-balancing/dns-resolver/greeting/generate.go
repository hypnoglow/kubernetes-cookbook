package greeting

//go:generate protoc -I. --go_out=plugins=grpc:$GOPATH/src greeter.proto
