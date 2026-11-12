# Lab 1: basic HTTP file server

[Repository overview](../../README.md).

The current basic implementation supports GET and POST for local files, with
at most ten active request handlers. It uses raw TCP networking and closes each
connection after one response. POST creates a file if absent and **appends** to
an existing file. Pub-sub and cloud bonus work are not implemented yet.

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

The port is a required argument. If 8080 is occupied, choose another port, such
as 8081, and use that same port in the client commands below. Leave the server
running while using a second terminal. Stop it with **Ctrl+C**.

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
- [Knowledge base](../../knowledge/README.md): standalone explanations and diagrams.
- [Agent guide](AGENTS.md): assignment constraints and working rules.
