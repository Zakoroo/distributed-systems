# JSON, field visibility, and byte payloads

Parent: [Go fundamentals](go-fundamentals.md).

JSON is a text format for values such as objects, arrays, strings, and numbers.
A response such as `{"id":1}` is an object with one property; an array of objects
looks like `[{"id":1},{"id":2}]`.

Go's `encoding/json` maps Go values to JSON with `Marshal`, and reads JSON with
`Unmarshal` or a streaming `Decoder`. Encoding and decoding can return errors.
Struct field tags choose property names:

```go
type subscriptionResponse struct {
    ID uint `json:"id"`
}
```

The field name must be exported for the encoder to include it. The tag controls
its JSON spelling, not Go access visibility. This does not require exposing a
service's internal fields: construct a separate response value containing only
what the client should receive. The response type itself can stay private.

JSON has no raw byte-string type. Go encodes `[]byte` using base64, a text
representation of bytes: the bytes for `hello` become `"aGVsbG8="`. This is why
an API's data property may differ from the text that was published. A non-nil
empty byte slice encodes as `""`; a nil byte slice encodes as `null`.

For a byte payload, decoding into a Go `[]byte` field reverses that encoding.
Document whether an API uses text, base64 bytes, or another format so clients can
interpret it correctly. See [Go's JSON package](https://pkg.go.dev/encoding/json).
