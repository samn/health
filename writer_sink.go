package health

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	ERROR
)

var logLevelToString = map[LogLevel]string{
	TRACE: "trace",
	DEBUG: "debug",
	INFO:  "info",
	ERROR: "error",
}

func (l LogLevel) String() string {
	return logLevelToString[l]
}

var stringToLogLevel = map[string]LogLevel{
	"trace": TRACE,
	"debug": DEBUG,
	"info":  INFO,
	"error": ERROR,
}

func (l *LogLevel) Scan(state fmt.ScanState, verb rune) error {
	token, err := state.Token(true, nil)
	if err != nil {
		return err
	}

	word := strings.ToLower(string(token))
	if level, ok := stringToLogLevel[word]; ok {
		*l = level
		return nil
	} else {
		return fmt.Errorf("No LogLevel found for %s", string(token))
	}
}

// This sink writes bytes in a format that a human might like to read in a logfile
// This can be used to log to Stdout:
//   .AddSink(WriterSink{os.Stdout})
// And to a file:
//   f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
//   .AddSink(WriterSink{f, INFO})
// And to syslog:
//   w, err := syslog.New(LOG_INFO, "wat")
//   .AddSink(WriterSink{w, INFO})
type WriterSink struct {
	io.Writer
	Level LogLevel
}

func (s *WriterSink) shouldLogEvent(kvs map[string]string) bool {
	if level, ok := kvs["level"]; ok {
		var eventLevel LogLevel
		n, err := fmt.Sscan(level, &eventLevel)
		if n != 1 || err != nil {
			// we couldn't decode a log level from the event, default to logging unknowns
			return true
		}

		return eventLevel >= s.Level
	}
	return true
}

func (s *WriterSink) EmitEvent(job string, event string, kvs map[string]string) {
	if !s.shouldLogEvent(kvs) {
		return
	}

	var b bytes.Buffer
	b.WriteRune('[')
	b.WriteString(timestamp())
	b.WriteString("]: job:")
	b.WriteString(job)
	b.WriteString(" event:")
	b.WriteString(event)
	writeMapConsistently(&b, kvs)
	b.WriteRune('\n')
	s.Writer.Write(b.Bytes())
}

func (s *WriterSink) EmitEventErr(job string, event string, inputErr error, kvs map[string]string) {
	if !s.shouldLogEvent(kvs) {
		return
	}

	var b bytes.Buffer
	b.WriteRune('[')
	b.WriteString(timestamp())
	b.WriteString("]: job:")
	b.WriteString(job)
	b.WriteString(" event:")
	b.WriteString(event)
	b.WriteString(" err:")
	b.WriteString(inputErr.Error())
	writeMapConsistently(&b, kvs)
	b.WriteRune('\n')
	s.Writer.Write(b.Bytes())
}

func (s *WriterSink) EmitTiming(job string, event string, nanos int64, kvs map[string]string) {
	if !s.shouldLogEvent(kvs) {
		return
	}

	var b bytes.Buffer
	b.WriteRune('[')
	b.WriteString(timestamp())
	b.WriteString("]: job:")
	b.WriteString(job)
	b.WriteString(" event:")
	b.WriteString(event)
	b.WriteString(" time:")
	writeNanoseconds(&b, nanos)
	writeMapConsistently(&b, kvs)
	b.WriteRune('\n')
	s.Writer.Write(b.Bytes())
}

func (s *WriterSink) EmitComplete(job string, status CompletionStatus, nanos int64, kvs map[string]string) {
	if !s.shouldLogEvent(kvs) {
		return
	}

	var b bytes.Buffer
	b.WriteRune('[')
	b.WriteString(timestamp())
	b.WriteString("]: job:")
	b.WriteString(job)
	b.WriteString(" status:")
	b.WriteString(status.String())
	b.WriteString(" time:")
	writeNanoseconds(&b, nanos)
	writeMapConsistently(&b, kvs)
	b.WriteRune('\n')
	s.Writer.Write(b.Bytes())
}

func timestamp() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func writeMapConsistently(b *bytes.Buffer, kvs map[string]string) {
	if kvs == nil {
		return
	}
	keys := make([]string, 0, len(kvs))
	for k := range kvs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	keysLenMinusOne := len(keys) - 1

	b.WriteString(" kvs:[")
	for i, k := range keys {
		b.WriteString(k)
		b.WriteRune(':')
		b.WriteString(kvs[k])

		if i != keysLenMinusOne {
			b.WriteRune(' ')
		}
	}
	b.WriteRune(']')
}

func writeNanoseconds(b *bytes.Buffer, nanos int64) {
	switch {
	case nanos > 2000000:
		fmt.Fprintf(b, "%d ms", nanos/1000000)
	case nanos > 2000:
		fmt.Fprintf(b, "%d μs", nanos/1000)
	default:
		fmt.Fprintf(b, "%d ns", nanos)
	}
}
