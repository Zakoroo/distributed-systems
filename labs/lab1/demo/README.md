# Lab 1 demo deck

The deck follows the whiteboard agenda, with click reveals in the configuration
slide and speaker notes for the live demo. Open `slides.md` to edit it.

```sh
cd labs/lab1/demo
npm ci
npm run dev
```

Open `http://localhost:3030` for the slides or `http://localhost:3030/presenter`
for speaker notes. See the [Slidev guide](https://sli.dev/guide/) for controls.

Start the Go server separately using the [Lab 1 run instructions](../README.md).
On **Configure the local demos**, set the server address and port. The default
is HTTP on `127.0.0.1:8080`. Add or edit client aliases, source IPs, and source
ports. These settings carry forward to the live slides and are saved in the
browser for the same Slidev origin.

The local Slidev process opens the socket using the selected client's source IP
and port. Use an IP assigned to the presentation machine and a free source port.
Port 0 lets the OS pick a port. Use distinct fixed ports for concurrent requests,
and switch clients while Alice's poll waits. The responses show the actual local
socket address and port. Network forwarding can change the address the remote
server sees. Keep the development deck on localhost.

In each demo, choose **Plain text** or **File upload** under **Payload type**.
Switching between them keeps both your draft text and the selected file.
For file API requests, the server appends the payload at the selected path and
accepts `.html`, `.txt`, `.css`, `.gif`, `.jpg`,
and `.jpeg`. Choose a fresh supported filename before uploading a binary file.
On **Live pub-sub demo**, subscribe as Alice, start her poll, then switch to Bob and
publish text or any file to the same topic. Switch back to Alice for her
subscriber ID. GET responses and received events can be downloaded.

On **Configure SSH forwarding**, fill in the SSH host and user, the local tunnel
port, and the HTTP destination as seen from the SSH host. Run the generated
command in a terminal. **Use tunnel for deployment demo** fills the separate
deployment target. You can also enter a directly reachable HTTP or HTTPS
address on **Configure the deployment server**. The deployment check shares the
client list with the local demos.

`npm run build` produces a static deck in `dist/`. The queue illustration runs
in the static deck. Use the local Slidev development server for live HTTP requests.

## Rehearse concurrent file access

Run this in a separate terminal with the local server listening on port 8080.
It appends two complete lines to a fresh file. Their order can vary.

```sh
file="concurrent-$(date +%s).txt"
curl -fsS --data-binary $'Alice\n' "http://localhost:8080/$file" &
alice=$!
curl -fsS --data-binary $'Bob\n' "http://localhost:8080/$file" &
bob=$!
wait "$alice" && wait "$bob" && curl -fsS "http://localhost:8080/$file"
```

Use the output as a demonstration. The race-enabled checks in the Lab 1 suite
exercise simultaneous reads and appends and check for complete records.
