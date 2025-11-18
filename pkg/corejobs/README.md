# Core Jobs

This package defines job types used by the Openlane API (core) with [Riverqueue](https://riverqueue.com/). Due to cyclical dependency issues, any job that uses the openlane API, needs to be defined in this repo, and just configured in `riverboat`.


## Usage

1. Add new jobs to this directory in a new file, refer to
   the [upstream docs](https://riverqueue.com/docs#job-args-and-workers) for
   implementation details. The following is a stem job that could be copied to
   get you started.

   ```go
   package jobs

   import (
      "context"

      "github.com/riverqueue/river"
      "github.com/rs/zerolog/log"
   )

   // ExampleArgs for the example worker to process the job
   type ExampleArgs struct {
      // ExampleArg is an example argument
      ExampleArg string `json:"example_arg"`
   }

   // Kind satisfies the river.Job interface
   func (ExampleArgs) Kind() string { return "example" }

   // ExampleWorker does all sorts of neat stuff
   type ExampleWorker struct {
      river.WorkerDefaults[ExampleArgs]

      ExampleConfig
   }

   // ExampleConfig contains the configuration for the example worker
   type ExampleConfig struct {
      // DevMode is a flag to enable dev mode so we don't actually send millions of carrier pigeons
      DevMode bool `koanf:"devmode" json:"devmode" jsonschema:"description=enable dev mode" default:"true"`
   }

   // Work satisfies the river.Worker interface for the example worker
   func (w *ExampleConfig) Work(ctx context.Context, job *river.Job[ExampleArgs]) error {
      // do some work

      return nil
   }
   ```

1. Add a test for the new job. There are
   additional helper functions that can be used, see
   [river test helpers](https://riverqueue.com/docs/testing) for details.
1. Add a `test` job to `test/` directory by creating a new directory with a
   `main.go` function that will insert the job into the queue.
1. Workers should be registered in the [riverboat](https://github.com/theopenlane/riverboat)

### Test Jobs

Included in the `test/` directory are test jobs corresponding to the job types
in `pkg/jobs`.

1. Start the `riverboat` server using `task run-dev`
1. Run the test main, for example the `email`:

   ```bash
   go run test/email/main.go
   ```
1. This should insert the job successfully, it should be processed by `river`
   and the email should be added to `fixtures/email`


## Contributing

See the [contributing](.github/CONTRIBUTING.md) guide for more information.