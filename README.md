# grpcannon
Playing with Go and gRPC

## Usage

```
./grpcannon -proto helloworld.proto -call helloworld.Greeter.SayHello -d '{"name":"Bob"}' localhost:50051
```
