package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config for the redis client used to store key-value pairs
type Config struct {
	// Enabled to enable redis client in the server
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// Address is the host:port to connect to redis
	Address string `json:"address" koanf:"address" default:"localhost:6379"`
	// Name of the connecting client
	Name string `json:"name" koanf:"name" default:"openlane"`
	// Username to connect to redis
	Username string `json:"username" koanf:"username"`
	// Password, must match the password specified in the server configuration
	Password string `json:"password" koanf:"password"`
	// DB to be selected after connecting to the server, 0 uses the default
	DB int `json:"db" koanf:"db" default:"0"`
	// Dial timeout for establishing new connections, defaults to 5s
	DialTimeout time.Duration `json:"dialTimeout" koanf:"dialTimeout" default:"5s"`
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Supported values:
	//   - `0` - default timeout (3 seconds).
	//   - `-1` - no timeout (block indefinitely).
	//   - `-2` - disables SetReadDeadline calls completely.
	ReadTimeout time.Duration `json:"readTimeout" koanf:"readTimeout" default:"0"`
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.  Supported values:
	//   - `0` - default timeout (3 seconds).
	//   - `-1` - no timeout (block indefinitely).
	//   - `-2` - disables SetWriteDeadline calls completely.
	WriteTimeout time.Duration `json:"writeTimeout" koanf:"writeTimeout" default:"0"`
	// MaxRetries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int `json:"maxRetries" koanf:"maxRetries" default:"3"`
	// MinIdleConns is useful when establishing new connection is slow.
	// Default is 0. the idle connections are not closed by default.
	MinIdleConns int `json:"minIdleConns" koanf:"minIdleConns" default:"0"`
	// Maximum number of idle connections.
	// Default is 0. the idle connections are not closed by default.
	MaxIdleConns int `json:"maxIdleConns" koanf:"maxIdleConns" default:"0"`
	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	MaxActiveConns int `json:"maxActiveConns" koanf:"maxActiveConns" default:"0"`
}

// New returns a new redis client based on the configuration settings
func New(c Config) *redis.Client {
	opts := &redis.Options{
		Addr:             c.Address,
		ClientName:       c.Name,
		DB:               c.DB,
		DialTimeout:      c.DialTimeout,
		ReadTimeout:      c.ReadTimeout,
		WriteTimeout:     c.WriteTimeout,
		MaxRetries:       c.MaxRetries,
		MinIdleConns:     c.MinIdleConns,
		MaxIdleConns:     c.MaxIdleConns,
		MaxActiveConns:   c.MaxActiveConns,
		DisableIndentity: true, // # spellcheck: off
	}

	if c.Username != "" {
		opts.Username = c.Username
	}

	if c.Password != "" {
		opts.Password = c.Password
	}

	return redis.NewClient(opts)
}

// Healthcheck pings the client to check if the connection is working
func Healthcheck(c *redis.Client) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// check if its alive
		if err := c.Ping(ctx).Err(); err != nil {
			return err
		}

		return nil
	}
}
