# build server for ubuntu
go build -v seed.go

# build client for ubuntu
go build -v -o go-node-ffi.so -buildmode=c-shared go-node-ffi.go

# build server for MAC
go build -v seed.go

# build client for MAC
cd client
go build -o go-node-ffi.dylib -buildmode=c-shared go-node-ffi.go

## CommKey

util/utils 中一个64位16进制随机数。用于客户端与服务器验证。

## 运行

./seed 