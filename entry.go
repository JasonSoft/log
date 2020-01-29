package log

import (
	"fmt"
	stdlog "log"
	"os"
	"sort"
	"strings"
	"time"
)

// Fields represents a map of entry level data used for structured logging.
type Fields map[string]interface{}

// Names returns field names sorted.
// map is not
func (f Fields) Names() (v []string) {
	for k := range f {
		v = append(v, k)
	}

	sort.Strings(v)
	return
}

// Get field value by name.
func (f Fields) Get(name string) interface{} {
	return f[name]
}

// Entry defines a single log entry
type Entry struct {
	logger *logger
	start  time.Time

	Level     Level     `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Fields    Fields    `json:"fields"`
}

func newEntry(l *logger) Entry {
	e := Entry{}
	e.logger = l
	e.Fields = l.defaultFields
	return e
}

// Debug level message.
func (e Entry) Debug(msg string) {
	e.Level = DebugLevel
	e.Message = msg
	handler(e)
}

// Debugf level message.
func (e Entry) Debugf(msg string, v ...interface{}) {
	e.Level = DebugLevel
	e.Message = fmt.Sprintf(msg, v...)
	handler(e)
}

// Println Info level message.
func (e Entry) Println(msg string) {
	e.Level = InfoLevel
	e.Message = msg
	handler(e)
}

// Print Info level message.
func (e Entry) Print(msg string) {
	e.Level = InfoLevel
	e.Message = msg
	handler(e)
}

// Info level message.
func (e Entry) Info(msg string) {
	e.Level = InfoLevel
	e.Message = msg
	handler(e)
}

// Infof level message.
func (e Entry) Infof(msg string, v ...interface{}) {
	e.Level = InfoLevel
	e.Message = fmt.Sprintf(msg, v...)
	handler(e)
}

// Warn level message.
func (e Entry) Warn(msg string) {
	e.Level = WarnLevel
	e.Message = msg
	handler(e)
}

// Warnf level message.
func (e Entry) Warnf(msg string, v ...interface{}) {
	e.Level = WarnLevel
	e.Message = fmt.Sprintf(msg, v...)
	handler(e)
}

// Error level message.
func (e Entry) Error(msg string) {
	e.Level = ErrorLevel
	e.Message = msg
	handler(e)
}

// Errorf level message.
func (e Entry) Errorf(msg string, v ...interface{}) {
	e.Level = ErrorLevel
	e.Message = fmt.Sprintf(msg, v...)
	handler(e)
}

// Panic level message.
func (e Entry) Panic(msg string) {
	e.Level = PanicLevel
	e.Message = msg
	handler(e)
	os.Exit(1)
}

// Panicf level message.
func (e Entry) Panicf(msg string, v ...interface{}) {
	e.Level = PanicLevel
	e.Message = fmt.Sprintf(msg, v...)
	handler(e)
	os.Exit(1)
}

// Fatal level message.
func (e Entry) Fatal(msg string) {
	e.Level = FatalLevel
	e.Message = msg
	handler(e)
	os.Exit(1)
}

// Fatalf level message.
func (e Entry) Fatalf(msg string, v ...interface{}) {
	e.Level = FatalLevel
	e.Message = fmt.Sprintf(msg, v...)
	handler(e)
	os.Exit(1)
}

// WithField returns a new entry with the `key` and `value` set.
func (e Entry) WithField(key string, value interface{}) Entry {
	return e.WithFields(Fields{key: value})
}

// WithFields adds the provided fields to the current entry
func (e Entry) WithFields(fields Fields) Entry {
	newFields := Fields{}
	if e.Fields != nil {
		for k, val := range e.Fields {
			newFields[k] = val
		}
	}
	if fields != nil {
		for k, val := range fields {
			newFields[k] = val
		}
	}

	e.Fields = newFields
	return e
}

// WithError returns a new entry with the "error" set to `err`.
func (e Entry) WithError(err error) Entry {
	if err == nil {
		return e
	}
	return e.WithField("error", fmt.Sprintf("%+v", err))
}

// Trace returns a new entry with a Stop method to fire off
// a corresponding completion log, useful with defer.
func (e Entry) Trace(msg string) Entry {
	e.Message = msg
	e.start = time.Now().UTC()
	return e
}

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

func duration(d time.Duration) string {
	if d < day {
		return d.String()
	}

	var b strings.Builder

	if d >= year {
		years := d / year
		fmt.Fprintf(&b, "%dy", years)
		d -= years * year
	}

	days := d / day
	d -= days * day
	fmt.Fprintf(&b, "%dd%s", days, d)

	return b.String()
}

// Stop should be used with Trace, to fire off the completion message. When
// an `err` is passed the "error" field is set, and the log level is error.
func (e Entry) Stop() {
	e.WithField("duration", duration(time.Since(e.start))).Info(e.Message)
}

func handler(e Entry) {
	e.Timestamp = time.Now().UTC()

	e.logger.rwMutex.RLock()
	defer e.logger.rwMutex.RUnlock()

	var err error
	for _, h := range e.logger.handlers[e.Level] {
		err = h.Log(e)
		if err != nil {
			stdlog.Printf("error logging: %s", err.Error())
		}

	}
}
