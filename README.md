# Enola Holmes
This is [Sherlock](https://github.com/sherlock-project/sherlock)'s sister **Enola**, Modern shiny CLI tool written with Golang to help you: 🔎 Hunt down social media accounts by username across social networks

## Install 
Minimum `go1.26` required.
```bash
go install github.com/theyahya/enola/cmd/enola@latest
```

## Usage
```bash
enola {username}
enola {username} --site twitter
enola {username} --output ./results.json
```

<img alt="Enola demo" src="https://github.com/theyahya/enola/blob/main/examples/demo.gif" width="600" />

### Using Docker
Build the image
```bash
docker build -t enola .
```

Run
```bash
docker run --rm -it enola {username}
```

## Library usage

```go
ctx := context.Background()
e, err := enola.New()
if err != nil {
    log.Fatal(err)
}

results, err := e.SetSite("twitter").Check(ctx, "username")
if err != nil {
    log.Fatal(err)
}

for r := range results {
    fmt.Println(r.Name, r.URL, r.Found)
}
```

Options: `enola.WithHTTPClient`, `enola.WithConcurrency`, `enola.WithData`.

## Contributing
You can fork the repository, improve or fix some part of it and then send a pull requests. Or simply open and issue if there's a bug, or you have a feature in mind.

To add a new detection strategy, create a file in `internal/checker/` that implements `Detector` and registers itself in `init()` via `checker.Register("your_type", ...)`.

## License

This software is released under the [MIT](https://github.com/TheYahya/enola/blob/main/LICENSE) License.
