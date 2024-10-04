# jsonupdate

Apply actions defined as JSON strings on arbitrary objects

For example, having the following object:

```go
type Internal struct {
    data string
}
type MyData struct {
    internal Internal
    what int
}

obj := MyData {
    internal: Internal{
        data: "nope"
    },
    what: 5
}
```

And the update:

```json
{
    "field": "internal.data",
    "action": "set",
    "value": "newdata"
}
```

Would result in

```go
obj = MyData {
    internal: Internal{
        data: "newdata"
    },
    what: 5
}
```