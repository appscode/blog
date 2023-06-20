---
title: How to use grpcurl to test gRPC servers
date: "2023-06-20"
weight: 11
authors:
- Abdul Matin
tags:
- grpc
- grpcurl
- rpc
---

[gRPC](https://grpc.io) is a high performance, open source universal RPC framework from Google. gRPC servers use a binary encoding on the wire (protocol buffers, or "protobufs" for short). So they are basically impossible to interact with using regular curl. `grpcurl` is a command-line tool that lets you interact with gRPC servers. It's basically curl for gRPC servers.

## How to use grpcurl tool to test gRPC server

First we need to have `grpcurl` installed in our system. Please follow [this](https://github.com/fullstorydev/grpcurl) and make sure that you can access `grpcurl` from you command line. 

Now, we run a sample gRPC server. You may find some sample gRPC servers on [Github](https://github.com/grpc/grpc/tree/master/examples).  I have a grpc server running at 0.0.0.0:50051. Let's see how we can use `grpcurl` to test it. Please make sure that the gRPC server you running supports Reflection API service. 

1. List services the server provides:

```bash
> grpcurl  -plaintext localhost:50051 list
grpc.health.v1.Health
grpc.reflection.v1alpha.ServerReflection
helloworld.Greeter
```

2. Describe any of those services

```bash
> grpcurl  -plaintext localhost:50051 describe helloworld.Greeter
helloworld.Greeter is a service:
service Greeter {
  rpc SayHello ( .helloworld.HelloRequest ) returns ( .helloworld.HelloReply );
  rpc SayHelloAgain ( .helloworld.HelloRequest ) returns ( .helloworld.HelloReply );
  rpc SayHelloStreamReply ( .helloworld.HelloRequest ) returns ( stream .helloworld.HelloReply );
}
```

3. Invoke a RPC

```bash
> grpcurl -plaintext -d '{"name": "matin"}'  localhost:50051 helloworld.Greeter.SayHello
{
  "message": "Hello matin"
}
```
