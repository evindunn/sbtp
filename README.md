# SBTP

**S**imple **B**inary **T**ransfer **P**rotocol for streaming transports implemented in go.
It is meant to be as simple to use as HTTP but slimmer.

Data transfers have the following format
```
| ------------------------------------------- |
| SBTPvx  | NULL   | Content-Length | NULL    |
| 6 bytes | 1 byte | 8 bytes        | 1 bytes |
| ------------------------------------------- |
| Payload                                     |
| Content-Length bytes                        |
| ------------------------------------------- |
| EOT                                         |
| 1 byte                                      |
| ------------------------------------------- |
```

* `SBTPvx` is the protocol version spec, where x is the version number
* `Content-Length` is a big-endian uint64
* `Payload` is the data to be sent

Examples can be found in [examples](./examples).

---

## Packets

SBTP Packets are a container for a `[]byte` payload. Packets implement the 
[`io.ReaderFrom`](https://pkg.go.dev/io#ReaderFrom) and [`io.WriterTo`](https://pkg.go.dev/io#WriterTo) interfaces to 
make send/recv over a [`net.Conn`](https://pkg.go.dev/net#Conn) trivial.

## Server

The SBTP server is a struct containing convenience methods for starting and stopping an SBTP server given a
[`net.Listener`](https://pkg.go.dev/net#Listener) that implements a `SetDeadline` method. The server contains a slice
of request handlers that process a request/response pair in order.

## Client

The SBTP client contains convenience methods for managing a [`net.Conn`](https://pkg.go.dev/net#Conn) to an SBTP
server. After calling `Connect()` SBTP packets can be sent and received using the `Request()` method.
