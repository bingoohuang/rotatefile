package stdlog

import (
	"fmt"
	"github.com/bingoohuang/rotatefile"
	"io"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/silentred/gid"
)

type wrapper struct {
	Writer io.Writer
}

func SetCaller(l bool) {
	DefaultCaller = l
}

func SetLevel(l Level) {
	DefaultLevel = l
}

func init() {
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		if level, err := ParseLevel(env[0]); err == nil {
			SetLevel(level)
		}
	}

	debugging := strings.Contains(os.Args[0], "/Caches/JetBrains")
	SetCaller(rotatefile.EnvBool("LOG_CALLER", debugging))
}

var (
	DefaultLevel  = InfoLevel
	DefaultCaller = false
)

func (w wrapper) Write(p []byte) (n int, err error) {
	level, p, _ := parseLevelFromMsg(p)
	if level > DefaultLevel {
		return len(p), nil
	}

	buf := getBuffer()
	defer putBuffer(buf)

	buf = writeTime(buf)
	*buf = append(*buf, ' ')

	buf = writeInfo(buf, level)
	*buf = append(*buf, ' ')

	*buf = append(*buf, pid...)
	*buf = append(*buf, ' ', '-', '-', '-', ' ')

	buf = writeGid(buf)
	*buf = append(*buf, ' ')

	buf = writeCaller(buf)
	*buf = append(*buf, ' ', ':', ' ')

	buf = writeMsg(p, buf)
	return w.Writer.Write(*buf)
}

func writeCaller(buf *[]byte) *[]byte {
	*buf = append(*buf, '[')
	if !DefaultCaller {
		*buf = append(*buf, '-', ']')
		return buf
	}

	_, file, line, ok := runtime.Caller(4)
	if !ok {
		file = "???"
		line = 0
	}
	{
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
	}
	*buf = append(*buf, file...)
	*buf = append(*buf, ':')
	itoa(buf, int64(line), -1)
	*buf = append(*buf, ']')
	return buf
}

func writeMsg(p []byte, buf *[]byte) *[]byte {
	i := len(p) - 1
	for ; i >= 0; i-- {
		if p[i] != '\r' && p[i] != '\n' {
			break
		}
	}
	*buf = append(*buf, p[:i]...)
	*buf = append(*buf, '\n')
	return buf
}

func writeGid(buf *[]byte) *[]byte {
	*buf = append(*buf, '[')
	id := gid.Get()
	len0 := len(*buf)
	*buf = strconv.AppendInt(*buf, id, 10)
	len1 := len(*buf)
	for i := len1 - len0; i < 6; i++ {
		*buf = append(*buf, ' ')
	}
	*buf = append(*buf, ']')
	return buf
}

func writeInfo(buf *[]byte, level Level) *[]byte {
	*buf = append(*buf, '[')
	levelBytes, _ := level.MarshalText()
	*buf = append(*buf, levelBytes...)
	if diff := 5 - len(levelBytes); diff > 0 {
		*buf = append(*buf, ' ')
	}
	*buf = append(*buf, ']')
	return buf
}

func writeTime(buf *[]byte) *[]byte {
	t := time.Now()
	{
		year, month, day := t.Date()
		itoa(buf, int64(year), 4)
		*buf = append(*buf, '-')
		itoa(buf, int64(month), 2)
		*buf = append(*buf, '-')
		itoa(buf, int64(day), 2)
		*buf = append(*buf, ' ')
	}
	{
		hour, min, sec := t.Clock()
		itoa(buf, int64(hour), 2)
		*buf = append(*buf, ':')
		itoa(buf, int64(min), 2)
		*buf = append(*buf, ':')
		itoa(buf, int64(sec), 2)
		*buf = append(*buf, '.')
		itoa(buf, int64(t.Nanosecond()/1e6), 3)
	}

	return buf
}

var pid = []byte(strconv.Itoa(os.Getpid()))

var bufferPool = sync.Pool{New: func() any { return new([]byte) }}

func getBuffer() *[]byte {
	p := bufferPool.Get().(*[]byte)
	*p = (*p)[:0]
	return p
}

func putBuffer(p *[]byte) {
	// Proper usage of a sync.Pool requires each entry to have approximately
	// the same memory cost. To obtain this property when the stored type
	// contains a variably-sized buffer, we add a hard limit on the maximum buffer
	// to place back in the pool.
	//
	// See https://go.dev/issue/23199
	if cap(*p) > 64<<10 { // 64M
		*p = nil
	}
	bufferPool.Put(p)
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int64, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func NewLevelLog(w io.Writer) io.Writer {
	return &wrapper{
		Writer: w,
	}
}

var (
	// regLevelTip parses the log level tip. the following tip is supported:
	// T! for trace
	// D! for debug
	// I! for info
	// W! for warn
	// E! for error
	// F! for fatal
	// P! for panic
	regLevelTip = regexp.MustCompile(`\b[TDIWEFP]!`)
)

func ParseLevelByte(b byte) Level {
	switch b {
	case 'T':
		return TraceLevel
	case 'D':
		return DebugLevel
	case 'I':
		return InfoLevel
	case 'W':
		return WarnLevel
	case 'E':
		return ErrorLevel
	case 'F':
		return FatalLevel
	case 'P':
		return PanicLevel
	default:
		return InfoLevel
	}
}

func parseLevelFromMsg(msg []byte) (level Level, s []byte, foundLevelTag bool) {
	if l := regLevelTip.FindIndex(msg); len(l) > 0 {
		x, y := l[0], l[1]
		level = ParseLevelByte(msg[x])
		if level <= PanicLevel {
			fmt.Println()
		}
		s = clearLevelFromMsg(msg, x, y)
		return level, s, true
	}

	return InfoLevel, msg, false
}

func clearLevelFromMsg(s []byte, x, y int) []byte {
	for ; x >= 0 && s[x] == ' '; x-- {
	}
	for ; y < len(s) && s[y] == ' '; y++ {
	}

	z := s[:x]
	if x > 0 {
		z = append(z, ' ')
	}
	return append(z, s[y:]...)
}

// Level type
type Level uint32

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func (level Level) String() string {
	if b, err := level.MarshalText(); err == nil {
		return string(b)
	} else {
		return "unknown"
	}
}

// ParseLevel takes a string level and returns the Logrus log level constant.
func ParseLevel(lvl byte) (Level, error) {
	switch lvl {
	case 'p', 'P':
		return PanicLevel, nil
	case 'f', 'F':
		return FatalLevel, nil
	case 'e', 'E':
		return ErrorLevel, nil
	case 'w', 'W':
		return WarnLevel, nil
	case 'i', 'I':
		return InfoLevel, nil
	case 'd', 'D':
		return DebugLevel, nil
	case 't', 'T':
		return TraceLevel, nil
	}

	return InfoLevel, fmt.Errorf("not a valid logrus Level: %q", lvl)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (level *Level) UnmarshalText(text []byte) error {
	l, err := ParseLevel(text[0])
	if err != nil {
		return err
	}

	*level = l

	return nil
}

func (level Level) MarshalText() ([]byte, error) {
	switch level {
	case TraceLevel:
		return []byte("TRACE"), nil
	case DebugLevel:
		return []byte("DEBUG"), nil
	case InfoLevel:
		return []byte("INFO"), nil
	case WarnLevel:
		return []byte("WARN"), nil
	case ErrorLevel:
		return []byte("ERROR"), nil
	case FatalLevel:
		return []byte("FATAL"), nil
	case PanicLevel:
		return []byte("PANIC"), nil
	}

	return nil, fmt.Errorf("not a valid logrus level %d", level)
}

// AllLevels A constant exposing all logging levels
var AllLevels = []Level{
	PanicLevel,
	FatalLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
	TraceLevel,
}

// These are the different logging levels. You can set the logging level to log
// on your instance of logger, obtained with `logrus.New()`.
const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)
