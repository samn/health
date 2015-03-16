package health

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

var basicEventRegexp = regexp.MustCompile("\\[[^\\]]+\\]: job:(.+) event:(.+)")
var kvsEventRegexp = regexp.MustCompile("\\[[^\\]]+\\]: job:(.+) event:(.+) kvs:\\[(.+)\\]")
var basicEventErrRegexp = regexp.MustCompile("\\[[^\\]]+\\]: job:(.+) event:(.+) err:(.+)")
var kvsEventErrRegexp = regexp.MustCompile("\\[[^\\]]+\\]: job:(.+) event:(.+) err:(.+) kvs:\\[(.+)\\]")
var basicTimingRegexp = regexp.MustCompile("\\[[^\\]]+\\]: job:(.+) event:(.+) time:(.+)")
var kvsTimingRegexp = regexp.MustCompile("\\[[^\\]]+\\]: job:(.+) event:(.+) time:(.+) kvs:\\[(.+)\\]")
var basicCompletionRegexp = regexp.MustCompile("\\[[^\\]]+\\]: job:(.+) status:(.+) time:(.+)")
var kvsCompletionRegexp = regexp.MustCompile("\\[[^\\]]+\\]: job:(.+) status:(.+) time:(.+) kvs:\\[(.+)\\]")

var testErr = errors.New("my test error")

func BenchmarkWriterSinkEmitEvent(b *testing.B) {
	var by bytes.Buffer
	someKvs := map[string]string{"foo": "bar", "qux": "dog"}
	sink := WriterSink{&by, TRACE}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		by.Reset()
		sink.EmitEvent("myjob", "myevent", someKvs)
	}
}

func BenchmarkWriterSinkEmitEventErr(b *testing.B) {
	var by bytes.Buffer
	someKvs := map[string]string{"foo": "bar", "qux": "dog"}
	sink := WriterSink{&by, TRACE}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		by.Reset()
		sink.EmitEventErr("myjob", "myevent", testErr, someKvs)
	}
}

func BenchmarkWriterSinkEmitTiming(b *testing.B) {
	var by bytes.Buffer
	someKvs := map[string]string{"foo": "bar", "qux": "dog"}
	sink := WriterSink{&by, TRACE}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		by.Reset()
		sink.EmitTiming("myjob", "myevent", 234203, someKvs)
	}
}

func BenchmarkWriterSinkEmitComplete(b *testing.B) {
	var by bytes.Buffer
	someKvs := map[string]string{"foo": "bar", "qux": "dog"}
	sink := WriterSink{&by, TRACE}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		by.Reset()
		sink.EmitComplete("myjob", Success, 234203, someKvs)
	}
}

func TestWriterSinkEmitEventBasic(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, INFO}
	sink.EmitEvent("myjob", "myevent", nil)

	str := b.String()

	result := basicEventRegexp.FindStringSubmatch(str)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
}

func TestWriterSinkEmitEventKvsWithFilteredLogLevel(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, INFO}
	kvs := map[string]string{
		"level": DEBUG.String(),
	}
	sink.EmitEvent("myjob", "myevent", kvs)

	assert.Equal(t, 0, b.Len())
}

func TestWriterSinkEmitEventKvsWithIncludedLogLevel(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, INFO}
	kvs := map[string]string{
		"level": ERROR.String(),
	}
	sink.EmitEvent("myjob", "myevent", kvs)

	str := b.String()

	result := kvsEventRegexp.FindStringSubmatch(str)
	assert.Equal(t, 4, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
	assert.Equal(t, "level:error", result[3])
}

func TestWriterSinkEmitEventKvs(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, INFO}
	sink.EmitEvent("myjob", "myevent", map[string]string{"wat": "ok", "another": "thing"})

	str := b.String()

	result := kvsEventRegexp.FindStringSubmatch(str)
	assert.Equal(t, 4, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
	assert.Equal(t, "another:thing wat:ok", result[3])
}

func TestWriterSinkEmitEventKvsWithLogLevel(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, INFO}
	sink.EmitEvent("myjob", "myevent", map[string]string{"wat": "ok", "another": "thing", "level": INFO.String()})

	str := b.String()

	result := kvsEventRegexp.FindStringSubmatch(str)
	assert.Equal(t, 4, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
	assert.Equal(t, "another:thing level:info wat:ok", result[3])
}

func TestWriterSinkEmitEventKvsAndFilteredLogLevel(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, ERROR}
	sink.EmitEvent("myjob", "myevent", map[string]string{"wat": "ok", "another": "thing", "level": INFO.String()})

	assert.Equal(t, 0, b.Len())
}

func TestWriterSinkEmitEventErrBasic(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, ERROR}
	sink.EmitEventErr("myjob", "myevent", testErr, nil)

	str := b.String()

	result := basicEventErrRegexp.FindStringSubmatch(str)
	assert.Equal(t, 4, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
	assert.Equal(t, testErr.Error(), result[3])
}

func TestWriterSinkEmitEventErrKvs(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, ERROR}
	sink.EmitEventErr("myjob", "myevent", testErr, map[string]string{"wat": "ok", "another": "thing"})

	str := b.String()

	result := kvsEventErrRegexp.FindStringSubmatch(str)
	assert.Equal(t, 5, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
	assert.Equal(t, testErr.Error(), result[3])
	assert.Equal(t, "another:thing wat:ok", result[4])
}

func TestWriterSinkEmitEventErrKvsWithLogLevel(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, INFO}
	sink.EmitEventErr("myjob", "myevent", testErr, map[string]string{"wat": "ok", "another": "thing", "level": INFO.String()})

	str := b.String()

	result := kvsEventErrRegexp.FindStringSubmatch(str)
	assert.Equal(t, 5, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
	assert.Equal(t, testErr.Error(), result[3])
	assert.Equal(t, "another:thing level:info wat:ok", result[4])
}

func TestWriterSinkEmitEventErrKvsAndFilteredLogLevel(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, ERROR}
	sink.EmitEventErr("myjob", "myevent", testErr, map[string]string{"wat": "ok", "another": "thing", "level": INFO.String()})

	assert.Equal(t, 0, b.Len())
}

func TestWriterSinkEmitTimingBasic(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, TRACE}
	sink.EmitTiming("myjob", "myevent", 1204000, nil)

	str := b.String()

	result := basicTimingRegexp.FindStringSubmatch(str)
	assert.Equal(t, 4, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
	assert.Equal(t, "1204 μs", result[3])
}

func TestWriterSinkEmitTimingKvs(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, ERROR}
	sink.EmitTiming("myjob", "myevent", 34567890, map[string]string{"wat": "ok", "another": "thing"})

	str := b.String()

	result := kvsTimingRegexp.FindStringSubmatch(str)
	assert.Equal(t, 5, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
	assert.Equal(t, "34 ms", result[3])
	assert.Equal(t, "another:thing wat:ok", result[4])
}

func TestWriterSinkEmitTimingKvsWithLogLevel(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, INFO}
	sink.EmitTiming("myjob", "myevent", 34567890, map[string]string{"wat": "ok", "another": "thing", "level": INFO.String()})

	str := b.String()

	result := kvsTimingRegexp.FindStringSubmatch(str)
	assert.Equal(t, 5, len(result))
	assert.Equal(t, "myjob", result[1])
	assert.Equal(t, "myevent", result[2])
	assert.Equal(t, "34 ms", result[3])
	assert.Equal(t, "another:thing level:info wat:ok", result[4])
}

func TestWriterSinkEmitTimingKvsAndFilteredLogLevel(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, ERROR}
	sink.EmitTiming("myjob", "myevent", 34567890, map[string]string{"wat": "ok", "another": "thing", "level": INFO.String()})

	assert.Equal(t, 0, b.Len())
}

func TestWriterSinkEmitCompleteBasic(t *testing.T) {
	for kind, kindStr := range completionStatusToString {
		var b bytes.Buffer
		sink := WriterSink{&b, ERROR}
		sink.EmitComplete("myjob", kind, 1204000, nil)

		str := b.String()

		result := basicCompletionRegexp.FindStringSubmatch(str)
		assert.Equal(t, 4, len(result))
		assert.Equal(t, "myjob", result[1])
		assert.Equal(t, kindStr, result[2])
		assert.Equal(t, "1204 μs", result[3])
	}
}

func TestWriterSinkEmitCompleteKvs(t *testing.T) {
	for kind, kindStr := range completionStatusToString {
		var b bytes.Buffer
		sink := WriterSink{&b, ERROR}
		sink.EmitComplete("myjob", kind, 34567890, map[string]string{"wat": "ok", "another": "thing"})

		str := b.String()

		result := kvsCompletionRegexp.FindStringSubmatch(str)
		assert.Equal(t, 5, len(result))
		assert.Equal(t, "myjob", result[1])
		assert.Equal(t, kindStr, result[2])
		assert.Equal(t, "34 ms", result[3])
		assert.Equal(t, "another:thing wat:ok", result[4])
	}
}

func TestWriterSinkEmitCompleteKvsWithLogLevel(t *testing.T) {
	for kind, kindStr := range completionStatusToString {
		var b bytes.Buffer
		sink := WriterSink{&b, INFO}
		sink.EmitComplete("myjob", kind, 34567890, map[string]string{"wat": "ok", "another": "thing", "level": INFO.String()})

		str := b.String()

		result := kvsCompletionRegexp.FindStringSubmatch(str)
		assert.Equal(t, 5, len(result))
		assert.Equal(t, "myjob", result[1])
		assert.Equal(t, kindStr, result[2])
		assert.Equal(t, "34 ms", result[3])
		assert.Equal(t, "another:thing level:info wat:ok", result[4])
	}
}

func TestWriterSinkEmitCompleteKvsAndFilteredLogLevel(t *testing.T) {
	for kind := range completionStatusToString {
		var b bytes.Buffer
		sink := WriterSink{&b, ERROR}
		sink.EmitComplete("myjob", kind, 34567890, map[string]string{"wat": "ok", "another": "thing", "level": INFO.String()})

		assert.Equal(t, 0, b.Len())
	}
}

func TestWriterSinkShouldLogEventTrace(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, TRACE}

	kvs := map[string]string{
		"level": TRACE.String(),
	}
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = DEBUG.String()
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = INFO.String()
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = ERROR.String()
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = ""
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = "wat"
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
}

func TestWriterSinkShouldLogEventDebug(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, DEBUG}

	kvs := map[string]string{
		"level": TRACE.String(),
	}
	assert.Equal(t, false, sink.shouldLogEvent(kvs))
	kvs["level"] = DEBUG.String()
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = INFO.String()
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = ERROR.String()
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = ""
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = "wat"
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
}

func TestWriterSinkShouldLogEventInfo(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, INFO}

	kvs := map[string]string{
		"level": TRACE.String(),
	}
	assert.Equal(t, false, sink.shouldLogEvent(kvs))
	kvs["level"] = DEBUG.String()
	assert.Equal(t, false, sink.shouldLogEvent(kvs))
	kvs["level"] = INFO.String()
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = ERROR.String()
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = ""
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = "wat"
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
}

func TestWriterSinkShouldLogEventError(t *testing.T) {
	var b bytes.Buffer
	sink := WriterSink{&b, ERROR}

	kvs := map[string]string{
		"level": TRACE.String(),
	}
	assert.Equal(t, false, sink.shouldLogEvent(kvs))
	kvs["level"] = DEBUG.String()
	assert.Equal(t, false, sink.shouldLogEvent(kvs))
	kvs["level"] = INFO.String()
	assert.Equal(t, false, sink.shouldLogEvent(kvs))
	kvs["level"] = ERROR.String()
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = ""
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
	kvs["level"] = "wat"
	assert.Equal(t, true, sink.shouldLogEvent(kvs))
}
