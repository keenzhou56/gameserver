package stdlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gameserver/pkg/log"
	"io"
	stdlog "log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var _ log.Logger = (*Logger)(nil)

// Option is std logger option.
type Option func(*options)

type options struct {
	prefix string
	flag   int
	skip   int
	out    io.WriteCloser
}

// Prefix with logger prefix.
func Prefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

// Flag with logger flag.
func Flag(flag int) Option {
	return func(o *options) {
		o.flag = flag
	}
}

// Skip with caller skip.
func Skip(skip int) Option {
	return func(o *options) {
		o.skip = skip
	}
}

// Writer with logger writer.
func Writer(out io.WriteCloser) Option {
	return func(o *options) {
		o.out = out
	}
}

// Logger is std logger.
type Logger struct {
	opts options
	log  *stdlog.Logger
	pool *sync.Pool
}

// NewLogger new a std logger with options.
func NewLogger(opts ...Option) *Logger {
	options := options{
		flag: stdlog.LstdFlags,
		skip: 4,
		out:  os.Stdout,
	}
	for _, o := range opts {
		o(&options)
	}
	return &Logger{
		opts: options,
		log:  stdlog.New(options.out, options.prefix, options.flag),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (s *Logger) stackTrace(path string) string {
	idx := strings.LastIndexByte(path, '/')
	if idx == -1 {
		return path
	}
	idx = strings.LastIndexByte(path[:idx], '/')
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}

// Print print the kv pairs log.
func (s *Logger) Print(kvpair ...interface{}) {
	if len(kvpair) == 0 {
		return
	}
	if len(kvpair)%2 != 0 {
		kvpair = append(kvpair, "")
	}

	buf := s.pool.Get().(*bytes.Buffer)
	if _, file, line, ok := runtime.Caller(s.opts.skip); ok {
		buf.WriteString(fmt.Sprintf("source=%s:%d ", s.stackTrace(file), line))
	}
	for i := 0; i < len(kvpair); i += 2 {
		fmt.Fprintf(buf, "%s=%v ", kvpair[i], kvpair[i+1])
	}
	s.log.Println(buf.String())
	buf.Reset()
	s.pool.Put(buf)
}

// Close close the logger.
func (s *Logger) Close() error {
	return s.opts.out.Close()
}

// PrintJSON print the kv pairs log.
func (s *Logger) PrintJSON(kvpair ...interface{}) {
	if len(kvpair) == 0 {
		return
	}
	if len(kvpair)%2 != 0 {
		kvpair = append(kvpair, "")
	}

	buf := s.pool.Get().(*bytes.Buffer)
	if _, file, line, ok := runtime.Caller(s.opts.skip); ok {
		buf.WriteString(fmt.Sprintf("source=%s:%d ", s.stackTrace(file), line))
	}
	jsonMap := make(map[string]interface{})
	for i := 0; i < len(kvpair); i += 2 {
		// fmt.Fprintf(buf, "%s=%v ", kvpair[i], kvpair[i+1])
		k := Strval(kvpair[i])
		jsonMap[k] = kvpair[i+1]
	}
	buf2, _ := json.Marshal(jsonMap)
	s.log.Println(buf.String(), string(buf2))
	buf.Reset()
	s.pool.Put(buf)
}

// Strval ....
func Strval(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}
