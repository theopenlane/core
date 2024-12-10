# Updates

https://github.com/theopenlane/core/blob/7d2331ae17a3c62d2301cc041ab0ad46909ea2b9/internal/httpserve/serveropts/option.go#L241

https://github.com/theopenlane/core/blob/7d2331ae17a3c62d2301cc041ab0ad46909ea2b9/pkg/middleware/debug/bodydump.go#L27

https://github.com/theopenlane/core/blob/5e27384d4789837ae610d0e1b722fcd89a9ee69f/internal/httpserve/serveropts/hooks.go#L8

https://github.com/theopenlane/core/blob/5e27384d4789837ae610d0e1b722fcd89a9ee69f/internal/graphapi/tools_test.go#L69

https://github.com/theopenlane/core/blob/5e27384d4789837ae610d0e1b722fcd89a9ee69f/internal/httpserve/handlers/tools_test.go#L62

https://github.com/theopenlane/core/blob/main/cmd/root.go#L62

https://github.com/theopenlane/core/blob/main/cmd/cli/cmd/root.go#L99

## Useful zerolog information for the future

### Sampling

```go
func samplerExample() {
	infoSampler := &zerolog.BurstSampler{
		Burst:  3,
		Period: 1 * time.Second,
	}

	warnSampler := &zerolog.BurstSampler{
		Burst:       3,
		Period:      1 * time.Second,
		NextSampler: &zerolog.BasicSampler{N: 5}, // Log every 5th message after exceeding the burst rate of 3 messages per
		// second
	}

	errorSampler := &zerolog.BasicSampler{N: 2}

	l := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Sample(zerolog.LevelSampler{
			WarnSampler:  warnSampler,
			InfoSampler:  infoSampler,
			ErrorSampler: errorSampler,
		})

	for i := 1; i <= 10; i++ {
		l.Info().Msgf("a message from the gods: %d", i)
		l.Warn().Msgf("warn message: %d", i)
		l.Error().Msgf("error message: %d", i)
	}
}
```