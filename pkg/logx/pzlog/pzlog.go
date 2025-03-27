package pzlog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/goccy/go-json"
	"github.com/gookit/color"
	"github.com/pterm/pterm"
	"github.com/rs/zerolog"
)

// Field represents a key-value pair in the log event
type Field struct {
	Key string
	Val string
}

// Event represents a log event with a timestamp, level, message, and fields
type Event struct {
	Timestamp string
	Level     string
	Message   string
	Fields    []Field
}

// Formatter is a function type that formats a value as a string
type Formatter func(interface{}) string

// PtermWriter is a custom writer for zerolog that formats log events using pterm
type PtermWriter struct {
	MaxWidth int
	Out      io.Writer

	LevelStyles map[zerolog.Level]*pterm.Style

	Tmpl *template.Template

	DefaultKeyStyle     func(string, zerolog.Level) *pterm.Style
	DefaultValFormatter func(string, zerolog.Level) Formatter

	KeyStyles     map[string]*pterm.Style
	ValFormatters map[string]Formatter

	KeyOrderFunc func(string, string) bool
}

// ErrCannotParseTemplate is an error indicating that a template cannot be parsed
var ErrCannotParseTemplate = errors.New("cannot parse template")

// NewPtermWriter creates a new PtermWriter with the provided options
func NewPtermWriter(options ...func(*PtermWriter)) *PtermWriter {
	pt := PtermWriter{
		MaxWidth: pterm.GetTerminalWidth(),
		Out:      os.Stdout,
		LevelStyles: map[zerolog.Level]*pterm.Style{
			zerolog.TraceLevel: pterm.NewStyle(pterm.Bold, pterm.FgCyan),
			zerolog.DebugLevel: pterm.NewStyle(pterm.Bold, pterm.FgBlue),
			zerolog.InfoLevel:  pterm.NewStyle(pterm.Bold, pterm.FgGreen),
			zerolog.WarnLevel:  pterm.NewStyle(pterm.Bold, pterm.FgYellow),
			zerolog.ErrorLevel: pterm.NewStyle(pterm.Bold, pterm.FgRed),
			zerolog.FatalLevel: pterm.NewStyle(pterm.Bold, pterm.FgRed),
			zerolog.PanicLevel: pterm.NewStyle(pterm.Bold, pterm.FgRed),
			zerolog.NoLevel:    pterm.NewStyle(pterm.Bold, pterm.FgWhite),
		},
		KeyStyles: map[string]*pterm.Style{
			zerolog.MessageFieldName:    pterm.NewStyle(pterm.Bold, pterm.FgWhite),
			zerolog.TimestampFieldName:  pterm.NewStyle(pterm.Bold, pterm.FgGray),
			zerolog.CallerFieldName:     pterm.NewStyle(pterm.Bold, pterm.FgGray),
			zerolog.ErrorFieldName:      pterm.NewStyle(pterm.Bold, pterm.FgRed),
			zerolog.ErrorStackFieldName: pterm.NewStyle(pterm.Bold, pterm.FgRed),
		},
		ValFormatters: map[string]Formatter{},
		KeyOrderFunc: func(k1, k2 string) bool {
			score := func(s string) string {
				s = color.ClearCode(s)
				if s == zerolog.TimestampFieldName {
					return string([]byte{0, 0})
				}
				if s == zerolog.CallerFieldName {
					return string([]byte{0, 1})
				}
				if s == zerolog.ErrorFieldName {
					return string([]byte{math.MaxUint8, 0})
				}
				if s == zerolog.ErrorStackFieldName {
					return string([]byte{math.MaxUint8, 1})
				}
				return s
			}
			return score(k1) < score(k2)
		},
	}

	tmpl := `{{ .Timestamp }} {{ .Level }}  {{ .Message }}
{{- range $i, $field := .Fields }}
{{ space (totalLength 1 $.Timestamp $.Level) }}{{if (last $i $.Fields )}}└{{else}}├{{ end }} {{ .Key }}: {{ .Val }}
{{- end }}
`
	t, err := template.New("event").
		Funcs(template.FuncMap{
			"space": func(n int) string {
				return strings.Repeat(" ", n)
			},
			"totalLength": func(n int, s ...string) int {
				return len(color.ClearCode(strings.Join(s, ""))) - n
			},
			"last": func(x int, a interface{}) bool {
				return x == reflect.ValueOf(a).Len()-1
			},
		}).
		Parse(tmpl)

	if err != nil {
		panic(fmt.Errorf("%w: %s", ErrCannotParseTemplate, err))
	}

	pt.Tmpl = t

	pt.DefaultKeyStyle = func(_ string, lvl zerolog.Level) *pterm.Style {
		return pt.LevelStyles[lvl]
	}

	pt.DefaultValFormatter = func(_ string, _ zerolog.Level) Formatter {
		return func(v any) string {
			return pterm.Sprint(v)
		}
	}

	for _, option := range options {
		option(&pt)
	}

	return &pt
}

func (pw *PtermWriter) Write(p []byte) (n int, err error) {
	return pw.Out.Write(p)
}

func (pw *PtermWriter) WriteLevel(lvl zerolog.Level, p []byte) (n int, err error) {
	var evt map[string]interface{}

	d := json.NewDecoder(bytes.NewReader(p))

	d.UseNumber()

	err = d.Decode(&evt)
	if err != nil {
		return n, fmt.Errorf("cannot decode event: %w", err)
	}

	var event Event
	if ts, ok := evt[zerolog.TimestampFieldName]; ok {
		event.Timestamp = pw.KeyStyles[zerolog.TimestampFieldName].Sprint(ts)
	}

	event.Level = pw.LevelStyles[lvl].Sprint(lvl)

	if msg, ok := evt[zerolog.MessageFieldName]; ok {
		event.Message = pw.KeyStyles[zerolog.MessageFieldName].Sprint(msg)
	}

	event.Fields = make([]Field, 0, len(evt))

	for k, v := range evt {
		if k == zerolog.TimestampFieldName ||
			k == zerolog.LevelFieldName ||
			k == zerolog.MessageFieldName {
			continue
		}

		var key string

		if style, ok := pw.KeyStyles[k]; ok {
			key = style.Sprint(k)
		} else {
			key = pw.DefaultKeyStyle(k, lvl).Sprint(k)
		}

		var val string
		if fn, ok := pw.ValFormatters[k]; ok {
			val = fn(v)
		} else {
			val = pw.DefaultValFormatter(k, lvl)(v)
		}

		event.Fields = append(event.Fields, Field{Key: key, Val: val})
	}

	sort.Slice(event.Fields, func(i, j int) bool {
		return pw.KeyOrderFunc(event.Fields[i].Key, event.Fields[j].Key)
	})

	var buf bytes.Buffer

	err = pw.Tmpl.Execute(&buf, event)
	if err != nil {
		return n, fmt.Errorf("cannot execute template: %w", err)
	}

	return pw.Out.Write(buf.Bytes())
}
