# opm

Go web framework

## [Guide]

### Installation

```sh
go get github.com/boyfinal/opm
```

### Example

```go
package main

import (
  "net/http"
  "github.com/boyfinal/opm"
  "github.com/boyfinal/opm/middleware"
)

func main() {
  // Opm instance
  or := opm.NewRouter()

  // Middleware
  or.Use(middleware.Logger())
  or.Use(middleware.Recover())

  // Routes
  or.GET("/", opm.HandlerFunc(func(c opm.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

  srv := &opm.Server{
		Addr:         "localhost:8888",
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      or,
	}

  // Start server
  srv.Run()
}
```
