# Lab 1: HTTP file server and small pub-sub broker

[Repository overview](../../README.md).

The current basic implementation supports GET and POST for local files, with
at most ten active request handlers. It uses raw TCP networking and closes each
connection after one response. POST creates a file if absent and **appends** to
an existing file. The small pub-sub broker shares the executable and port. Event
polls count toward the same ten-handler limit. Cloud deployment remains deferred.

## Build and start

First complete the [workspace setup](../../README.md#setup). Then, in a terminal
at the repository root:

```sh
cd labs/lab1
mkdir -p public
go build -o http_server ./src
./http_server 8080
```

If already inside `labs/lab1`, skip the `cd` command. Run the executable from this
directory: `./public/` is relative to the **working directory**, not the source or
binary location. The server does not create that directory itself. Both `public/`
and the compiled executable are ignored by Git.

For a terminal run, pass the port as shown above. If 8080 is occupied, choose
another port, such as 8081, and use it in the client commands below. Leave the server
running while using a second terminal. Stop it with **Ctrl+C**.

## Run in Docker

From the repository root, build the binary and then the image:

```sh
cd labs/lab1
go build -o http_server ./src
docker build -t http_server .
docker run --rm -p 8080:80 http_server
```

The image copies the compiled binary. Rebuild both after Go source changes.
The server also accepts `PORT` when no command-line port is supplied; the
Dockerfile sets it to 80. `-p 8080:80` forwards host port 8080 to container port 80.
Stop any old container already using host port 8080 before starting the new one.
Rebuilding an image does not change a container that is already running.

The Dockerfile sets `WORKDIR /app` and creates `/app/public`. Consequently,
`POST /hello.txt` writes `/app/public/hello.txt` inside the container. Creating
`public/` on the host alone does not create it inside the image. See
[working directories and container filesystems](../../knowledge/concepts/files/files-and-paths.md#working-directories-in-containers).

Try the running container from another terminal:

```sh
curl -i --data-binary 'Hello world!' http://localhost:8080/hello.txt
curl -i http://localhost:8080/hello.txt
```

Files stay in this container's filesystem unless storage is mounted separately;
removing the container removes those files. The `--rm` option removes this demo
container when it exits.

## Use curl as a client

Run these commands in a **second terminal**, from any directory. `-i` includes
the response headers so you can see the HTTP status. `--data-binary` sends the
given bytes as a POST request body; without a body option, curl uses GET here.

### Create and read a file

```sh
curl -i --data-binary 'Hello' http://127.0.0.1:8080/readme-demo.txt
curl -i http://127.0.0.1:8080/readme-demo.txt
```

Both requests should return **200 OK**. If `readme-demo.txt` did not already
exist, GET returns `Hello` with `Content-Type: text/plain`. The server stores the
file at `labs/lab1/public/readme-demo.txt`.

### Append more content

```sh
curl -i --data-binary ', world!' http://127.0.0.1:8080/readme-demo.txt
curl -i http://127.0.0.1:8080/readme-demo.txt
```

With a new demo file and the preceding commands run once, the body is now
`Hello, world!`. Repeating POST appends again; it does not reset the file.

To send an existing local file instead of a literal string, use `@` before its
path, running curl from the directory containing that client-side file:

```sh
curl -i --data-binary @notes.txt http://127.0.0.1:8080/upload.txt
```

This also appends if `upload.txt` already exists on the server. For nested server
paths, the parent directories must already exist under `public/`.

### Try error responses

```sh
curl -i http://127.0.0.1:8080/missing-demo.txt
curl -i http://127.0.0.1:8080/readme-demo.json
curl -i -X PUT http://127.0.0.1:8080/readme-demo.txt
```

| Request | Expected status |
| --- | --- |
| GET `missing-demo.txt`, assuming it does not exist | **404 Not Found** |
| GET a `.json` path, an unsupported extension | **400 Bad Request** |
| PUT, an unsupported method | **501 Not Implemented** |

Supported extensions are `.html`, `.txt`, `.gif`, `.jpeg`, `.jpg`, and `.css`.
Invalid filenames/paths and other file-operation failures return 400. The
[assignment instructions](docs/lab1_instructions.pdf) define the required behavior.

## Layout

```text
lab1/
├── AGENTS.md             Agent guidance
├── go.mod                Module definition
├── README.md             Build/run/client instructions
├── src/                  Executable, tests, and application packages
│   ├── main.go
│   ├── interfaces/fileapi/
│   ├── interfaces/pubsubapi/  One broker file and its tests
│   ├── middleware/
│   └── filepolicy/
├── docs/                 Assignment, design, and learning notes
└── public/               Local runtime data, created during setup
```

## Tests and further reading

From `labs/lab1`, run the application's checks with `go test -race ./...`.
For shared-module checks too, use the commands in the
[root README](../../README.md#check-the-implementation).

- [Learning map](docs/LEARNING.md): reading order, concepts, and source links.
- [Server design](docs/DESIGN.md): package boundaries and request lifecycle.
- [Broker walkthrough](docs/PUBSUB.md): subscriber queues, message publication,
  waiting, unsubscribe, and the ordinary handler limit.
- [Knowledge base](../../knowledge/README.md): standalone explanations and diagrams.
- [Agent guide](AGENTS.md): assignment constraints and working rules.

## Pub-sub client example

The broker uses the same executable and port. Subscribe in a client terminal:

```sh
curl -i -X POST 'http://127.0.0.1:8080/subscribe?topic=weather'
```

The response is 201 with an ID such as `{"id":1}`. Use the actual returned ID.
Ask for one event in that terminal:

```sh
curl -i 'http://127.0.0.1:8080/events?topic=weather&id=1'
```

If no event is queued, this request waits. Publish from another terminal:

```sh
curl -i --data-binary 'rain' 'http://127.0.0.1:8080/publish?topic=weather'
```

Publication returns 200. The poll returns 200 with one JSON object such as
`{"sequence":1,"data":"cmFpbg=="}`; the data is base64 for `rain`. JSON property
order does not matter. Poll again after finishing the response. An empty wait
returns 204 after the configured 15-second timeout; it does not unsubscribe you.

```sh
curl -i -X POST 'http://127.0.0.1:8080/unsubscribe?topic=weather&id=1'
```

Unsubscribe returns 200; later polls for that ID return 404. Clients keep one
outstanding poll per ID and read its complete response before polling again.
Messages and subscriptions are in memory. Each queue holds 64 events and skips
new messages when full. **Waiting polls count toward the ordinary ten-handler
limit:** ten waiting polls can delay a publisher until a handler finishes.
See the [walkthrough](docs/PUBSUB.md) for the code and these deliberate limits.
