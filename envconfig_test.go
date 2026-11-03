package envconfig

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

type database struct {
	URL      string `env:"URL"`
	MaxConns int    `env:"MAX_CONNS" default:"10"`
}

type config struct {
	Port     string        `env:"PORT" default:"8080"`
	Debug    bool          `env:"DEBUG"`
	Timeout  time.Duration `env:"TIMEOUT" default:"5s"`
	Hosts    []string      `env:"HOSTS"`
	Database database      `env:"DB"`
	Cache    *database     `env:"CACHE"`
	internal string
}

func TestLoad(t *testing.T) {
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_DEBUG", "true")
	t.Setenv("APP_HOSTS", "a, b,,c")
	t.Setenv("APP_DB_URL", "postgres://localhost/app")
	t.Setenv("APP_CACHE_MAX_CONNS", "3")

	var c config
	if err := Load("APP", &c); err != nil {
		t.Fatal(err)
	}

	want := config{
		Port:     "9090",
		Debug:    true,
		Timeout:  5 * time.Second,
		Hosts:    []string{"a", "b", "c"},
		Database: database{URL: "postgres://localhost/app", MaxConns: 10},
		Cache:    &database{MaxConns: 3},
	}
	if !reflect.DeepEqual(c, want) {
		t.Fatalf("got %+v, want %+v", c, want)
	}
}

func TestLoadNoPrefix(t *testing.T) {
	t.Setenv("PORT", "3000")

	var c struct {
		Port string `env:"PORT" default:"8080"`
	}
	if err := Load("", &c); err != nil {
		t.Fatal(err)
	}
	if c.Port != "3000" {
		t.Fatalf("got %q", c.Port)
	}
}

func TestLoadInvalidValue(t *testing.T) {
	t.Setenv("APP_DB_MAX_CONNS", "lots")

	var c config
	err := Load("APP", &c)
	if err == nil || err.Error() != "envconfig: APP_DB_MAX_CONNS: not a valid int" {
		t.Fatalf("got %v", err)
	}
}

func TestLoadInvalidValueNotInError(t *testing.T) {
	t.Setenv("APP_TIMEOUT", "s3cr3t-token")

	var c config
	err := Load("APP", &c)
	if err == nil || strings.Contains(err.Error(), "s3cr3t-token") {
		t.Fatalf("got %v", err)
	}
}

func TestLoadRejectsNonFiniteFloat(t *testing.T) {
	for _, raw := range []string{"NaN", "+Inf", "-Inf"} {
		t.Run(raw, func(t *testing.T) {
			t.Setenv("APP_RATIO", raw)

			var c struct {
				Ratio float64 `env:"RATIO"`
			}
			err := Load("APP", &c)
			if err == nil || err.Error() != "envconfig: APP_RATIO: not a valid float64" {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestLoadNotStructPointer(t *testing.T) {
	var c config
	if err := Load("APP", c); err == nil {
		t.Fatal("expected an error")
	}
}

type node struct {
	Name string `env:"NAME"`
	Next *node  `env:"NEXT"`
}

func TestLoadSelfReferential(t *testing.T) {
	var n node
	err := Load("APP", &n)
	if err == nil || !strings.Contains(err.Error(), "structs nested deeper than 32") {
		t.Fatalf("got %v", err)
	}
}
