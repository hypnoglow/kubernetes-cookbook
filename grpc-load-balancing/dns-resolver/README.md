# gRPC load balancing using default DNS resolver

TODO: add detailed description.

## Caveats

Default DNS resolver is a simple solution but it has a problem. If it 
happens that all your gRPC server replicas will be down/unavailable, your gRPC client may not reconnect
to the server when it becomes available (client may be stuck for up to 30 minutes).
See [this issue](https://github.com/grpc/grpc-go/issues/1795)
for details.

## Run examples

    skaffold run
