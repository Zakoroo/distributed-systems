# TCP and connections

Parent: [Networking and HTTP](networking-and-http.md).

**TCP** gives two programs an ordered stream of bytes. The program opening a
connection is the client; the program listening for connections is the server.
Neither term says which machine is more powerful.

An **IP address** identifies a network destination. A **port** helps select a
service at that destination. An address such as `127.0.0.1:8080` combines both;
`127.0.0.1` is local loopback. Listening on `:8080` does not restrict the listener
to one explicitly named local interface.

## Listener and connection are different objects

```go
listener, err := net.Listen("tcp", ":8080")
// Check err before continuing.
conn, err := listener.Accept()
// Check err before reading or writing conn.
```

The listener waits for new connections. Each successful `Accept` returns a
separate `net.Conn` used to exchange bytes with one client. Closing the listener
does not automatically close connections already accepted from it.

Reads may wait for bytes, and writes may wait for the peer or network. Such a
call is **blocking**. With ordinary Go network connections, a waiting goroutine
can be parked while other goroutines run. That waiter still holds its open
connection, memory, and any application admission slot it owns.

## Receiving bytes is not receiving a complete message

One read may contain part of a request, one request, or bytes from more than one
request. TCP preserves byte order, not the boundaries between application writes.
[HTTP](http-basics.md) supplies message structure; [buffering](../files/streams-and-buffering.md)
helps a parser consume the stream.

`SetReadDeadline` and `SetWriteDeadline` bound how long I/O may wait, including
already-blocked operations. A zero deadline means no timeout. A deadline is an
absolute time; an idle-time policy may need to update it deliberately.

The application must choose who closes each connection. Connection reuse needs
more than leaving a socket open: both peers must agree on where messages end.
See [Go's net.Conn documentation](https://pkg.go.dev/net#Conn).
