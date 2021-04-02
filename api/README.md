# Api

## Usage

### Structured proto

```
.\bin\protoc.exe --proto_path=./src/im/api/protocol/ --go_out=./ protocol.proto
.\bin\protoc.exe --proto_path=./src/im/api/logic/ --go-grpc_out=./ --go_out=./ logic.proto
```