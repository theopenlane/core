# echolog

[Zerolog](https://github.com/rs/zerolog) wrapper for [Echo](https://echo.labstack.com/) web framework.

## Installation

```
go get github.com/theopenlane/core/pkg/logx/echolog
```


## Quick start

```go
package main

import (
	"os"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
	"github.com/theopenlane/core/pkg/logx/echolog"
)

func main() {
    e := echo.New()
    e.Logger = echolog.New(os.Stdout)
}
```

### Using existing zerolog instance

```go
package main

import (
	"os"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
	"github.com/theopenlane/core/pkg/logx/echolog"
    "github.com/rs/zerolog"
)

func main() {
    log := zerolog.New(os.Stdout)
    e := echo.New()
    e.Logger = echolog.From(log)
}

```

## Options

```go

import (
	"os",
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
	"github.com/theopenlane/core/pkg/logx/echolog"
)

func main() {
    e := echo.New()
    e.Logger = echolog.New(
       os.Stdout,
       echolog.WithLevel(log.DEBUG),
       echolog.WithFields(map[string]interface{}{ "name": "hot diggity dogs"}),
       echolog.WithTimestamp(),
       echolog.WithCaller(),
       echolog.WithPrefix("❤️ MITB"),
       echolog.WithHook(...),
       echolog.WithHookFunc(...),
    )
}
```

## Middleware

### Logging requests and attaching request id to a context logger

```go

import (
	"os",
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
	"github.com/theopenlane/core/pkg/logx/echolog"
	"github.com/rs/zerolog"
)

func main() {
    e := echo.New()
    logger := echolog.New(
            os.Stdout,
            echolog.WithLevel(log.DEBUG),
            echolog.WithTimestamp(),
            echolog.WithCaller(),
         )
    e.Logger = logger

    e.Use(middleware.RequestID())
    e.Use(echolog.Middleware(echolog.Config{
    	Logger: logger
    }))
    e.GET("/", func(c echo.Context) error {
        c.Logger().Print("echos interface")
        zerolog.Ctx(c.Request().Context()).Print("zerlogs interface")

	return c.String(http.StatusOK, "MITB babbyyyyy")
    })
}

```

### Escalate log level for slow requests:

```go
e.Use(echolog.Middleware(echolog.Config{
    Logger: logger,
    RequestLatencyLevel: zerolog.WarnLevel,
    RequestLatencyLimit: 500 * time.Millisecond,
}))
```


### Nesting under a sub dictionary

```go
e.Use(echolog.Middleware(echolog.Config{
        Logger: logger,
        NestKey: "request"
    }))
    // Output: {"level":"info","request":{"remote_ip":"5.6.7.8","method":"GET", ...}, ...}
```

### Enricher

Enricher allows you to add additional fields to the log entry.

```go
e.Use(echolog.Middleware(echolog.Config{
        Logger: logger,
        Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
            return e.Str("user_id", c.Get("user_id"))
        },
    }))
    // Output: {"level":"info","user_id":"123", ...}
```

```go
Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
  userId := c.Get("UserID")
  if userId != nil {
    return logger.Str("user_id", userId.(string))
  }
  return logger.Str("user_id", "")
}
```

### Errors

The middleware does not automatically propagate errors up the chain.  If you want to do that, you can set `HandleError` to ``true``.

```go
e.Use(echolog.Middleware(echolog.Config{
    Logger: logger,
    HandleError: true,
}))
```

## Helpers

### Level converters

```go

import (
    "fmt"
    echo "github.com/theopenlane/echox"
    "github.com/theopenlane/echox/middleware"
    "github.com/labstack/gommon/log"
    "github.com/theopenlane/core/pkg/logx/echolog"
)

func main() {
	var z zerolog.Level
	var e log.Lvl

    z, e = echolog.MatchEchoLevel(log.WARN)

    fmt.Println(z, e)

    e, z = echolog.MatchZeroLevel(zerolog.INFO)

    fmt.Println(z, e)
}

```

## Multiple Log output

echolog.New(zerolog.MultiLevelWriter(consoleWriter, os.Stdout))

echolog.From(zerolog.New(zerolog.MultiLevelWriter(consoleWriter, os.Stdout)))

echolog.Middleware(echolog.Config{
    Logger: echolog.New(zerolog.MultiLevelWriter(consoleWriter, os.Stdout)),
})

### Writing to a file with Lumberjack

echolog.New(&lumberjack.Logger{
    Filename:   "/var/log/myapp/foo.log",
    MaxSize:    500, // megabytes
    MaxBackups: 3,
    MaxAge:     28, //days
    Compress:   true, // disabled by default
})