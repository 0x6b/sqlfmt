# sqlfmt

A naive Go library that provides SQL formatting capabilities by using the [sql-formatter](https://github.com/sql-formatter-org/sql-formatter/) JavaScript library through [go-quickjs](https://github.com/rosbit/go-quickjs).

## Installation

To use `sqlfmt` in your Go project, you can simply import it:

```go
import "github.com/0x6b/sqlfmt"
```

Then, run `go mod tidy` to download the dependencies.

## Usage

The `sqlfmt` package exposes a `FormatSQL` function and a `DefaultFormatOptions` variable. You can use `DefaultFormatOptions` and override specific fields as needed. See [example](examples/main.go) for usage.

## Acknowledgements

The `assets` directory contains `sql-formatter.min.js` (version 15.6.6), which is an artifact from the [sql-formatter](https://github.com/sql-formatter-org/sql-formatter) project.

## License

MIT. See [LICENSE](LICENSE) for details.
