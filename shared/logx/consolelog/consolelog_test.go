package consolelog_test

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/theopenlane/shared/logx/consolelog"
)

func TestConsolelog(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		w := consolelog.NewConsoleWriter()

		if w.TimeFormat == "" {
			t.Errorf("Missing w.TimeFormat")
		}

		if w.Formatter("foobar") == nil {
			t.Errorf(`Missing default formatter for "foobar"`)
		}

		d := time.Unix(0, 0).UTC().Format(time.RFC3339)
		o := w.Formatter("time")(d)
		if o != "12:00AM" {
			t.Errorf(`Unexpected output for date %q: %s`, d, o)
		}
	})

	t.Run("SetFormatter", func(t *testing.T) {
		w := consolelog.NewConsoleWriter()

		w.SetFormatter("time", func(i interface{}) string { return "FOOBAR" })

		d := time.Unix(0, 0).UTC().Format(time.RFC3339)
		o := w.Formatter("time")(d)
		if o != "FOOBAR" {
			t.Errorf(`Unexpected output from custom "time" formatter: %s`, o)
		}
	})

	t.Run("New with options", func(t *testing.T) {
		w := consolelog.NewConsoleWriter(
			func(w *consolelog.ConsoleWriter) {
				w.SetFormatter("time", func(i interface{}) string { return "FOOBAR" })
			},
		)

		d := time.Unix(0, 0).UTC().Format(time.RFC3339)
		o := w.Formatter("time")(d)
		if o != "FOOBAR" {
			t.Errorf(`Unexpected output from custom "time" formatter: %s`, o)
		}
	})

	t.Run("Write", func(t *testing.T) {
		var out bytes.Buffer
		w := consolelog.NewConsoleWriter()
		w.Out = &out

		d := time.Unix(0, 0).UTC().Format(time.RFC3339)
		_, err := w.Write([]byte(`{"time" : "` + d + `", "level" : "info", "message" : "Foobar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "12:00AM INF Foobar\n"
		actualOutput := out.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Write fields", func(t *testing.T) {
		var out bytes.Buffer
		w := consolelog.NewConsoleWriter()
		w.Out = &out

		d := time.Unix(0, 0).UTC().Format(time.RFC3339)
		_, err := w.Write([]byte(`{"time" : "` + d + `", "level" : "debug", "message" : "Foobar", "foo" : "bar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "12:00AM DBG Foobar foo=bar\n"
		actualOutput := out.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Write caller", func(t *testing.T) {
		var out bytes.Buffer
		w := consolelog.NewConsoleWriter()
		w.Out = &out

		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Cannot get working directory: %s", err)
		}

		d := time.Unix(0, 0).UTC().Format(time.RFC3339)
		evt := `{"time" : "` + d + `", "level" : "debug", "message" : "Foobar", "foo" : "bar", "caller" : "` + cwd + `/foo/bar.go"}`
		// t.Log(evt)

		_, err = w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "12:00AM DBG foo/bar.go | Foobar foo=bar\n"
		actualOutput := out.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Write error", func(t *testing.T) {
		var out bytes.Buffer
		w := consolelog.NewConsoleWriter()
		w.Out = &out

		d := time.Unix(0, 0).UTC().Format(time.RFC3339)
		evt := `{"time" : "` + d + `", "level" : "error", "message" : "Foobar", "aaa" : "bbb", "error" : "Error"}`
		// t.Log(evt)

		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "12:00AM ERR Foobar error=Error aaa=bbb\n"
		actualOutput := out.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})
}

func BenchmarkConsolelog(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	var msg = []byte(`{"level" : "info", "foo" : "bar"}`)

	w := consolelog.NewConsoleWriter()
	w.Out = io.Discard

	for i := 0; i < b.N; i++ {
		w.Write(msg)
	}
}
