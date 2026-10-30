# envconfig

Fills Go structs from environment variables. Shared by the Harbour Commerce services.

```go
type Config struct {
	Port    string        `env:"PORT" default:"8080"`
	Timeout time.Duration `env:"TIMEOUT" default:"5s"`
	DB      struct {
		URL string `env:"URL"`
	} `env:"DB"`
}

var cfg Config
if err := envconfig.Load("", &cfg); err != nil {
	log.Fatal(err)
}
```

A field reads the variable named by its `env` tag, or its upper-cased name, joined to the prefix with `_`. `default` is used when the variable is unset. Nested structs extend the prefix, so `DB.URL` above reads `DB_URL`.

Supported types: `string`, `bool`, integers, floats, `time.Duration`, and `[]string` (comma-separated).

```sh
go test ./...
```
