# drpc

A simple RPC framework implemented in golang

### 1. 谈谈 RPC 框架

RPC(Remote Procedure Call，远程过程调用)是一种计算机通信协议，允许调用不同进程空间的程序。RPC 的客户端和服务器可以在一台机器上，也可以在不同的机器上。程序员使用时，就像调用本地程序一样，无需关注内部的实现细节。

不同的应用程序之间的通信方式有很多，比如浏览器和服务器之间广泛使用的基于 HTTP 协议的 Restful API。与 RPC 相比，Restful API 有相对统一的标准，因而更通用，兼容性更好，支持不同的语言。HTTP
协议是基于文本的，一般具备更好的可读性。但是缺点也很明显：

Restful 接口需要额外的定义，无论是客户端还是服务端，都需要额外的代码来处理，而 RPC 调用则更接近于直接调用。 基于 HTTP 协议的 Restful 报文冗余，承载了过多的无效信息，而 RPC
通常使用自定义的协议格式，减少冗余报文。 RPC 可以采用更高效的序列化协议，将文本转为二进制传输，获得更高的性能。 因为 RPC 的灵活性，所以更容易扩展和集成诸如注册中心、负载均衡等功能。

### 2. RPC 框架需要解决什么问题

RPC 框架需要解决什么问题？或者我们换一个问题，为什么需要 RPC 框架？

我们可以想象下两台机器上，两个应用程序之间需要通信，那么首先，需要确定采用的传输协议是什么？如果这个两个应用程序位于不同的机器，那么一般会选择 TCP 协议或者 HTTP 协议；那如果两个应用程序位于相同的机器，也可以选择 Unix
Socket 协议。传输协议确定之后，还需要确定报文的编码格式，比如采用最常用的 JSON 或者 XML，那如果报文比较大，还可能会选择 protobuf
等其他的编码方式，甚至编码之后，再进行压缩。接收端获取报文则需要相反的过程，先解压再解码。

解决了传输协议和报文编码的问题，接下来还需要解决一系列的可用性问题，例如，连接超时了怎么办？是否支持异步请求和并发？

如果服务端的实例很多，客户端并不关心这些实例的地址和部署位置，只关心自己能否获取到期待的结果，那就引出了注册中心(registry)和负载均衡(load balance)
的问题。简单地说，即客户端和服务端互相不感知对方的存在，服务端启动时将自己注册到注册中心，客户端调用时，从注册中心获取到所有可用的实例，选择一个来调用。这样服务端和客户端只需要感知注册中心的存在就够了。注册中心通常还需要实现服务动态添加、删除，使用心跳确保服务处于可用状态等功能。

再进一步，假设服务端是不同的团队提供的，如果没有统一的 RPC 框架，各个团队的服务提供方就需要各自实现一套消息编解码、连接池、收发线程、超时处理等“业务之外”的重复技术劳动，造成整体的低效。因此，“业务之外”的这部分公共的能力，即是
RPC 框架所需要具备的能力。

### 3. 关于 drpc

Go 语言广泛地应用于云计算和微服务，成熟的 RPC 框架和微服务框架汗牛充栋。grpc、rpcx、go-micro 等都是非常成熟的框架。一般而言，RPC 是微服务框架的一个子集，微服务框架可以自己实现 RPC
部分，当然，也可以选择不同的 RPC 框架作为通信基座。

考虑性能和功能，上述成熟的框架代码量都比较庞大，而且通常和第三方库，例如 protobuf、etcd、zookeeper 等有比较深的耦合，难以直观地窥视框架的本质。GeeRPC 的目的是以最少的代码，实现 RPC
框架中最为重要的部分，帮助大家理解 RPC 框架在设计时需要考虑什么。代码简洁是第一位的，功能是第二位的。

因此，drpc 选择从零实现 Go 语言官方的标准库 net/rpc，并在此基础上，新增了协议交换(protocol exchange)、注册中心(registry)、服务发现(service discovery)、负载均衡(load
balance)、超时处理(timeout processing)等特性。分七天完成，最终代码约 1000 行。

> https://geektutu.com/post/geerpc.html


### 4. log
```
2021/05/23 17:06:32 rpc registry path:  /_drpc_/registry
2021/05/23 17:06:33 rpc server: register Foo.Sleep
2021/05/23 17:06:33 rpc server: register Foo.Sum
2021/05/23 17:06:33 rpc server: register Foo.Sleep
2021/05/23 17:06:33 rpc server: register Foo.Sum
2021/05/23 17:06:33 tcp@[::]:57815 send heart beat to registry http://localhost:9999/_drpc_/registry
2021/05/23 17:06:33 tcp@[::]:57814 send heart beat to registry http://localhost:9999/_drpc_/registry
2021/05/23 17:06:33 rpc registry: putServer=tcp@[::]:57814, Numbers=2
2021/05/23 17:06:33 rpc registry: putServer=tcp@[::]:57815, Numbers=1
2021/05/23 17:06:34 rpc registry: refresh servers from registry: http://localhost:9999/_drpc_/registry 0
2021/05/23 17:06:34 [tcp@[::]:57814 tcp@[::]:57815]
2021/05/23 17:06:34 call Foo.Sum success: 0 + 0 = 0
2021/05/23 17:06:34 call Foo.Sum success: 1 + 1 = 2
2021/05/23 17:06:34 call Foo.Sum success: 3 + 9 = 12
2021/05/23 17:06:34 call Foo.Sum success: 2 + 4 = 6
2021/05/23 17:06:34 call Foo.Sum success: 4 + 16 = 20
2021/05/23 17:06:34 rpc registry: refresh servers from registry: http://localhost:9999/_drpc_/registry 0
2021/05/23 17:06:34 [tcp@[::]:57814 tcp@[::]:57815]
2021/05/23 17:06:34 broadcast Foo.Sum success: 4 + 16 = 20
2021/05/23 17:06:34 broadcast Foo.Sum success: 0 + 0 = 0
2021/05/23 17:06:34 broadcast Foo.Sum success: 1 + 1 = 2
2021/05/23 17:06:34 broadcast Foo.Sum success: 3 + 9 = 12
2021/05/23 17:06:34 broadcast Foo.Sum success: 2 + 4 = 6
2021/05/23 17:06:34 broadcast Foo.Sleep success: 0 + 0 = 0
2021/05/23 17:06:35 broadcast Foo.Sleep success: 1 + 1 = 2
2021/05/23 17:06:36 broadcast Foo.Sleep error: rpc client: call failed: context deadline exceeded
2021/05/23 17:06:36 broadcast Foo.Sleep error: rpc client: call failed: context deadline exceeded
2021/05/23 17:06:36 broadcast Foo.Sleep error: rpc client: call failed: context deadline exceeded
2021/05/23 17:06:36 rpc server: read header error: read tcp [::1]:57815->[::1]:57820: read: connection reset by peer
2021/05/23 17:06:36 rpc server: read header error: read tcp [::1]:57814->[::1]:57821: read: connection reset by peer
```