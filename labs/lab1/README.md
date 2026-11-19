# Lab 1 · HTTP file server and pub-sub

One Go server handles file requests and the pub-sub bonus on the same port.
See the [assignment instructions](docs/lab1_instructions.pdf).

## Build and run

From this folder, with Bash and Go 1.27.3 or newer:

```sh
./setup.sh
./http_server 8080
```

Setup creates `public/` and builds the executable. Run the server from this folder
because `public/` is relative to its working directory. Go downloads the pinned
modules for a standalone copy and uses local modules when a workspace is active.

## Try it

In a second terminal, send two appends and read the file. Use a fresh filename to
start with an empty file.

```sh
curl -i --data-binary 'Hello' http://localhost:8080/demo.txt
curl -i --data-binary ' world' http://localhost:8080/demo.txt
curl -i http://localhost:8080/demo.txt
```

The GET returns `Hello world`. The server stores the file at `public/demo.txt`.
Use `/demo.txt` in the URL because the server already resolves paths inside `public/`.

For pub-sub, subscribe and use the returned ID in the next command:

```sh
curl -i -X POST 'http://localhost:8080/subscribe?topic=weather'
curl -i 'http://localhost:8080/events?topic=weather&id=1'
```

While the poll waits, publish from another terminal:

```sh
curl -i --data-binary 'rain' 'http://localhost:8080/publish?topic=weather'
```

The poll returns one JSON event. An empty poll returns 204 after 15 seconds.
To unsubscribe, POST to `/unsubscribe?topic=weather&id=1` with the same ID.

## Run in Docker

```sh
./setup.sh
docker build -t http_server .
docker run --rm -p 8080:80 http_server
```

Use the same curl commands. Stop any server already using port 8080 first.
Files are stored at `/app/public` in the container. Mount a volume there to keep
files after removing the container.

## Review and test

Run `go test -race ./...` and `go vet ./...` from this folder.

- [Lab 1 design](docs/DESIGN.md) covers routing, workers, file locking, and the broker.
- [Slidev demo](demo/README.md) has presentation and live demo instructions.
