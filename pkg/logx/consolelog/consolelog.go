package consolelog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"slices"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

const (
	defaultTimeFormat = time.Kitchen
)

var (
	bold   = color.New(color.Bold).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	faint  = color.New(color.Faint).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	orange = color.New(color.FgHiYellow).SprintFunc()

	defaultFormatter  = func(i any) string { return fmt.Sprintf("%v", i) }
	defaultPartsOrder = []string{
		zerolog.TimestampFieldName,
		zerolog.LevelFieldName,
		zerolog.CallerFieldName,
		zerolog.MessageFieldName,
	}
)

// ConsoleWriter parses the JSON input and writes an ANSI-colorized output to out
type ConsoleWriter struct {
	Out           io.Writer
	TimeFormat    string
	PartsOrder    []string
	formatters    map[string]Formatter
	FieldsExclude []string
}

// Formatter transforms the input into a string
type Formatter func(any) string

type event map[string]any

// NewConsoleWriter creates and initializes a new ConsoleWriter
func NewConsoleWriter(options ...func(w *ConsoleWriter)) ConsoleWriter {
	w := ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: defaultTimeFormat,
		PartsOrder: defaultPartsOrder,
		formatters: make(map[string]Formatter),
	}

	w.setDefaultFormatters()

	for _, opt := range options {
		opt(&w)
	}

	return w
}

// Formatter returns a formatter by id or the default formatter if none is found
func (w *ConsoleWriter) Formatter(id string) Formatter {
	if f, ok := w.formatters[id]; ok {
		return f
	}

	return defaultFormatter
}

// SetFormatter registers a formatter function by id
func (w *ConsoleWriter) SetFormatter(id string, f Formatter) {
	w.formatters[id] = f
}

// Write appends the output to Out.
func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	var buf bytes.Buffer

	var evt event

	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()

	err = d.Decode(&evt)
	if err != nil {
		return n, err
	}

	for _, p := range w.PartsOrder {
		w.writePart(&buf, evt, p)
	}

	w.writeFields(evt, &buf)

	buf.WriteByte('\n')

	if _, err := buf.WriteTo(w.Out); err != nil {
		return len(p), err
	}

	return len(p), nil
}

func (w *ConsoleWriter) writePart(buf *bytes.Buffer, evt event, p string) {
	var s = w.Formatter(p)(evt[p])
	if len(s) > 0 {
		buf.WriteString(s)

		if p != w.PartsOrder[len(w.PartsOrder)-1] {
			buf.WriteByte(' ')
		}
	}
}

func (w *ConsoleWriter) writeFields(evt event, buf *bytes.Buffer) {
	var fields = make([]string, 0, len(evt))

	for field := range evt {
		switch field {
		case zerolog.LevelFieldName, zerolog.TimestampFieldName, zerolog.MessageFieldName, zerolog.CallerFieldName:
			continue
		}

		fields = append(fields, field)
	}

	sort.Strings(fields)

	if len(fields) > 0 {
		buf.WriteByte(' ')
	}

	if slices.Contains(fields, zerolog.ErrorFieldName) {
		fields = append([]string{zerolog.ErrorFieldName}, fields...)
		for i := 1; i < len(fields); i++ {
			if fields[i] == zerolog.ErrorFieldName {
				fields = slices.Delete(fields, i, i+1)
				break
			}
		}
	}

	for i, field := range fields {
		var fn Formatter

		var fv Formatter

		if _, ok := w.formatters[field+"_field_name"]; ok {
			fn = w.Formatter(field + "_field_name")
			fv = w.Formatter(field + "_field_value")
		} else {
			fn = w.Formatter("field_name")
			fv = w.Formatter("field_value")
		}

		buf.WriteString(fn(field))
		buf.WriteString(fv(evt[field]))

		if i < len(fields)-1 {
			buf.WriteByte(' ')
		}
	}
}

func (w *ConsoleWriter) setDefaultFormatters() {
	w.SetFormatter(
		zerolog.TimestampFieldName,
		func(i any) string {
			var t string

			if tt, ok := i.(string); ok {
				ts, err := time.Parse(time.RFC3339, tt)
				if err != nil {
					t = tt
				} else {
					t = ts.Format(w.TimeFormat)
				}
			}

			return faint(t)
		})
	w.SetFormatter(
		zerolog.LevelFieldName,

		func(i any) string {
			var l string

			if ll, ok := i.(string); ok {
				switch ll {
				case "debug":
					l = yellow("DBG")
				case "info":
					l = green("INF")
				case "warn":
					l = red("WRN")
				case "error":
					l = red("ERR")
				case "fatal":
					l = bold(red("FTL"))
				case "panic":
					l = bold(red("PNC"))
				default:
					l = bold("N/A")
				}
			} else {
				l = strings.ToUpper(fmt.Sprintf("%v", i))[0:3]
			}

			return l
		})
	w.SetFormatter(
		zerolog.CallerFieldName,
		func(i any) string {
			var c string

			if cc, ok := i.(string); ok {
				c = cc
			}

			if len(c) > 0 {
				cwd, err := os.Getwd()
				if err == nil {
					c = strings.TrimPrefix(c, cwd)
					c = strings.TrimPrefix(c, "/")
				}

				c = faint(orange(c)) + faint(" |")
			}

			return c
		})
	// message
	w.SetFormatter(
		zerolog.MessageFieldName,
		func(i any) string { return fmt.Sprintf("%s", i) })
	// field name
	w.SetFormatter(
		"field_name", func(i any) string {
			return cyan(fmt.Sprintf("%s=", i))
		})
	// field value
	w.SetFormatter(
		"field_value", func(i any) string {
			return fmt.Sprintf("%s", i)
		})
	// errors
	w.SetFormatter(
		"error_field_name", func(i any) string {
			return faint(red(fmt.Sprintf("%s=", i)))
		})
	w.SetFormatter(
		"error_field_value", func(i any) string {
			return bold(red(fmt.Sprintf("%s", i)))
		})
}
