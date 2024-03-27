// Package rotatefile provides a rolling logger.
//
// rotatefile is intended to be one part of a logging infrastructure.
// It is not an all-in-one solution, but instead is a pluggable
// component at the bottom of the logging stack that simply controls the files
// to which logs are written.
//
// rotatefile plays well with any logging package that can write to an
// io.Writer, including the standard library's log package.
//
// rotatefile assumes that only one process is writing to the output files.
// Using the same rotatefile configuration from multiple processes on the same
// machine will result in improper behavior.
package rotatefile

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bingoohuang/q"
	"github.com/bingoohuang/rotatefile/disk"
	"github.com/bingoohuang/rotatefile/flock"
)

// file is an io.WriteCloser that writes to the specified filename.
//
// file opens or creates the logfile on first Write.  If the file exists and
// is less than MaxSize megabytes, rotatefile will open and append to that file.
// If the file exists and its size is >= MaxSize megabytes, the file is renamed
// by putting the current time in a timestamp in the name immediately before the
// file's extension (or the end of the filename if there's no extension). A new
// log file is then created using original filename.
//
// Whenever a write would cause the current log file exceed MaxSize megabytes,
// the current file is closed, renamed, and a new log file created with the
// original name. Thus, the filename you give file is always the "current" log
// file.
//
// Backups use the log file name given to file, in the form
// `name-timestamp.ext` where name is the filename without the extension,
// timestamp is the time at which the log was rotated formatted with the
// time.Time format of `20060102T150405.000` and the extension is the
// original extension.  For example, if your file.Filename is
// `/var/log/foo/server.log`, a backup created at 6:30pm on Nov 11 2016 would
// use the filename `/var/log/foo/server.20161104T183000.000.log`
//
// # Cleaning Up Old Log Files
//
// Whenever a new logfile gets created, old log files may be deleted.  The most
// recent files according to the encoded timestamp will be retained, up to a
// number equal to MaxBackups (or all of them if MaxBackups is 0).  Any files
// with an encoded timestamp older than MaxDays days are deleted, regardless of
// MaxBackups.  Note that the time encoded in the timestamp is the rotation
// time, which may differ from the last time that file was written to.
//
// If MaxBackups and MaxDays are both 0, no old log files will be deleted.
type file struct {
	file   *os.File
	millCh chan bool

	flock *flock.Flock

	dir      string
	filename string

	Config

	size      atomic.Int64
	startMill sync.Once
	mu        sync.Mutex
}

// RotateFile 滚动文件大小
type RotateFile interface {
	io.WriteCloser

	Rotate() error
	Flush() error

	// GetFilename 取得日志文件的距离路径
	GetFilename() string
}

// New 创建新一个新的滚动文件对象
func New(fns ...ConfigFn) RotateFile {
	return &file{Config: createConfig(fns...)}
}

var (
	// currentTime exists, so it can be mocked out by tests.
	currentTime = time.Now

	// os_Stat exists, so it can be mocked out by tests.
	osStat = os.Stat
)

// Write implements io.Writer.  If a White would cause the log file to be larger
// than MaxSize, the file is closed, renamed to include a timestamp of the
// current time, and a new log file is created using the original log file name.
// If the length of to write is greater than MaxSize, an error is returned.
func (l *file) Write(p []byte) (n int, err error) {
	n, err = l.writeInternal(p)
	if err != nil {
		q.Q(err)
	}
	return
}

func (l *file) writeInternal(p []byte) (n int, err error) {
	if l.PrintTerm {
		os.Stdout.Write(p)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	writeLen := int64(len(p))
	if writeLen > l.max() {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum file size %d", writeLen, l.max(),
		)
	}

	if l.file == nil {
		if err = l.openExistingOrNew(len(p)); err != nil {
			return 0, err
		}
	}

	if l.size.Load()+writeLen > l.max() {
		if err := l.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = l.file.Write(p)
	l.size.Add(int64(n))

	return n, err
}

// Flush 刷新文件缓存到磁盘
// 当写入 warn 级别以上日志时，建议写完后，Flush 刷盘
func (l *file) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Sync()
	}

	if l.PrintTerm {
		os.Stdout.Sync()
	}

	return nil
}

func (l *file) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.close()
}

// close closes the file if it is open.
func (l *file) close() error {
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}

// Rotate causes file to close the existing log file and immediately create a
// new one.  This is a helper function for applications that want to initiate
// rotations outside the normal rotation rules, such as in response to
// SIGHUP.  After rotating, this initiates compression and removal of old log
// files according to the configuration.
func (l *file) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rotate()
}

// rotate closes the current file, moves it aside with a timestamp in the name,
// (if it exists), opens a new file with the original filename, and then runs
// post-rotation processing and removal.
func (l *file) rotate() error {
	if err := l.close(); err != nil {
		return err
	}
	if err := l.openNew(); err != nil {
		return err
	}
	l.mill()
	return nil
}

// openNew opens a new log file for writing, moving any old log file out of the
// way.  These methods assume the file has already been closed.
func (l *file) openNew() error {
	err := os.MkdirAll(l.dir, 0o755)
	if err != nil {
		return fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	name := l.filename
	mode := os.FileMode(0o600)
	if info, err := osStat(name); err == nil && info != nil {
		// Copy the mode off the old logfile.
		mode = info.Mode()
		// move the existing file
		newName := backupName(name, l.UtcTime)
		if err := os.Rename(name, newName); err != nil {
			return fmt.Errorf("can't rename log file: %s", err)
		}

		// this is a no-op anywhere but linux
		if err := chown(name, info); err != nil {
			return err
		}
	}

	// we use truncate here because this should only get called when we've moved
	// the file ourselves. if someone else creates the file in the meantime,
	// just wipe out the contents.
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	l.file = f
	l.size.Store(0)
	return nil
}

// backupName creates a new filename from the given name, inserting a timestamp
// between the filename and the extension, using the local time if requested
// (otherwise UTC).
func backupName(name string, utc bool) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	t := currentTime()
	if utc {
		t = t.UTC()
	}

	timestamp := t.Format(backupTimeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s.%s%s", prefix, timestamp, ext))
}

// openExistingOrNew opens the logfile if it exists and if the current write
// would not put it over MaxSize.  If there is no such file or the write would
// put it over the MaxSize, a new file is created.
func (l *file) openExistingOrNew(writeLen int) error {
	l.mill()

	filename := l.filename
	info, err := osStat(filename)
	if os.IsNotExist(err) {
		return l.openNew()
	}
	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	var size int64
	if info != nil {
		size = info.Size()
	}

	if size+int64(writeLen) >= l.max() {
		return l.rotate()
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return l.openNew()
	}
	l.file = file
	l.size.Store(size)
	return nil
}

func (l *file) GetFilename() string {
	return l.filename
}

// setFileName generates the name of the logfile from the current time.
func (l *file) setFileName() {
	l.filename, l.flock = GenerateFilename(l.AppName, l.Prefix, l.Filename, true)
	l.dir = filepath.Dir(l.filename)
}

// GenerateFilename 根据 filename 生成完整的
// 1. filename 为 /some/path/xxx.log, 则继续保持
// 2. filename 为 /some/path/, 则补齐日志文件名为: {appName}{currentDirBase}.log
// 3. filename 为 空, 则根据 FindLogDir 生成指定的日志目录，日志文件名见上
func GenerateFilename(appName, prefix, filename string, tryLock bool) (string, *flock.Flock) {
	logDir, logName := filename, ""
	if strings.HasSuffix(filename, ".log") {
		// 配置的是具体的日志文件名称（推荐的配置）
		logDir = filepath.Dir(filename)
		logName = filepath.Base(filename)
	}

	// 否则当做日志路径看待，日志文件名自动补全
	return getLogFileName(appName, logDir, prefix, logName, tryLock)
}

// millRunOnce performs compression and removal of stale log files.
// Log files are compressed if enabled via configuration and old log
// files are removed, keeping at most l.MaxBackups files, as long as
// none of them are older than MaxDays.
func (l *file) millRunOnce() error {
	if l.MaxBackups == 0 && l.MaxDays == 0 && !l.Compress {
		return nil
	}

	files, err := l.oldLogFiles()
	if err != nil {
		return err
	}

	var compress, remove []logInfo

	if l.MaxBackups > 0 && l.MaxBackups < len(files) {
		preserved := make(map[string]bool)
		var remaining []logInfo
		for _, f := range files {
			// Only count the uncompressed log file or the
			// compressed log file, not both.
			fn := f.Name
			if strings.HasSuffix(fn, compressSuffix) {
				fn = fn[:len(fn)-len(compressSuffix)]
			}
			preserved[fn] = true

			if len(preserved) > l.MaxBackups {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}
	if l.MaxDays > 0 {
		diff := 24 * time.Hour * time.Duration(l.MaxDays)
		cutoff := currentTime().Add(-diff)

		var remaining []logInfo
		for _, f := range files {
			if f.timestamp.Before(cutoff) {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}

	if l.Compress {
		for _, f := range files {
			if !strings.HasSuffix(f.Name, compressSuffix) {
				compress = append(compress, f)
			}
		}
	}

	dir := l.dir
	for _, f := range remove {
		removeFile := filepath.Join(dir, f.Name)
		errRemove := os.Remove(removeFile)
		if err == nil && errRemove != nil {
			err = errRemove
		}
	}
	for _, f := range compress {
		fn := filepath.Join(dir, f.Name)
		errCompress := compressLogFile(fn, fn+compressSuffix)
		if err == nil && errCompress != nil {
			err = errCompress
		}
	}

	if errTotalSizeCap := l.keepTotalSizeCap(dir); errTotalSizeCap != nil && err == nil {
		err = errTotalSizeCap
	}

	return err
}

func (l *file) keepTotalSizeCap(dir string) error {
	var dirDiskFree uint64

	if l.MinDiskFree > 0 {
		if dirDisk, err := disk.GetInfo(dir, false); err == nil {
			dirDiskFree = dirDisk.Free
		}
	}

	if l.TotalSizeCap <= 0 && (l.MinDiskFree == 0 || dirDiskFree >= l.MinDiskFree) {
		return nil
	}

	files, err := l.oldLogFiles()
	if err != nil {
		return err
	}

	totalSize := l.size.Load()
	for _, f := range files {
		totalSize += f.Size
	}

	// 从最近的历史文件开始，删除历史文件，以控制总大小
	for i := len(files) - 1; i >= 0; i-- {
		if uint64(totalSize) <= l.TotalSizeCap && (l.MinDiskFree == 0 || dirDiskFree >= l.MinDiskFree) {
			break
		}

		f := files[i]
		if err1 := os.Remove(filepath.Join(dir, f.Name)); err1 == nil {
			// 删除成功，从总大小中减去删除文件的大小
			totalSize -= f.Size
			dirDiskFree += uint64(f.Size)
		} else if err == nil {
			err = err1
		}
	}

	return err
}

// millRun runs in a goroutine to manage post-rotation compression and removal
// of old log files.
func (l *file) millRun() {
	for range l.millCh {
		// what am I going to do, log this?
		_ = l.millRunOnce()
	}
}

// mill performs post-rotation compression and removal of stale log files,
// starting the mill goroutine if necessary.
func (l *file) mill() {
	l.startMill.Do(func() {
		l.setFileName()
		l.signalRotate()
		l.millCh = make(chan bool, 1)
		go l.millRun()
	})
	select {
	case l.millCh <- true:
	default:
	}
}

func (l *file) signalRotate() {
	if len(l.RotateSignals) == 0 {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, l.RotateSignals...)

	go func() {
		for range c {
			l.Rotate()
		}
	}()
}

// oldLogFiles returns the list of backup log files stored in the same
// directory as the current log file, sorted by ModTime
// 不包括当前正在写入的日志文件，排序从最新到最老
func (l *file) oldLogFiles() ([]logInfo, error) {
	files, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	var logFiles []logInfo

	prefix, ext := l.prefixAndExt()

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		size := int64(0)
		if info, _ := f.Info(); info != nil {
			size = info.Size()
		}

		if t, err := l.timeFromName(f.Name(), prefix, ext); err == nil {
			logFiles = append(logFiles, logInfo{timestamp: t, Name: f.Name(), Size: size})
			continue
		}
		if t, err := l.timeFromName(f.Name(), prefix, ext+compressSuffix); err == nil {
			logFiles = append(logFiles, logInfo{timestamp: t, Name: f.Name(), Size: size})
			continue
		}
		// error parsing means that the suffix at the end was not generated
		// by rotatefile, and therefore it's not a backup file.
	}

	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].timestamp.After(logFiles[j].timestamp)
	})

	return logFiles, nil
}

// ErrMismatched defines the error type for mismatched time format.
var ErrMismatched = errors.New("mismatched extension/prefix/backup time format")

// timeFromName extracts the formatted time from the filename by stripping off
// the filename's prefix and extension. This prevents someone's filename from
// confusing time.parse.
func (l *file) timeFromName(filename, prefix, ext string) (time.Time, error) {
	if !strings.HasSuffix(filename, ext) {
		return time.Time{}, ErrMismatched
	}
	filename = filename[:len(filename)-len(ext)]
	if !strings.HasPrefix(filename, prefix) {
		return time.Time{}, ErrMismatched
	}

	ts := filename[len(prefix):]
	if len(ts) != len(backupTimeFormat) {
		return time.Time{}, ErrMismatched
	}

	return time.Parse(backupTimeFormat, ts)
}

// max returns the maximum size in bytes of log files before rolling.
func (l *file) max() int64 {
	if l.MaxSize == 0 {
		return defaultMaxSize
	}
	return int64(l.MaxSize)
}

// prefixAndExt returns the filename part and extension part from the file's
// filename.
func (l *file) prefixAndExt() (prefix, ext string) {
	filename := filepath.Base(l.filename)
	ext = filepath.Ext(filename)
	prefix = filename[:len(filename)-len(ext)] + "."
	return prefix, ext
}

// compressLogFile compresses the given log file, removing the
// uncompressed log file if successful.
func compressLogFile(src, dst string) (err error) {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer f.Close()

	fi, err := osStat(src)
	if err != nil {
		return fmt.Errorf("failed to stat log file: %v", err)
	}

	if err := chown(dst, fi); err != nil {
		return fmt.Errorf("failed to chown compressed log file: %v", err)
	}

	// If this file already exists, we presume it was created by
	// a previous attempt to compress the log file.
	gzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
	if err != nil {
		return fmt.Errorf("failed to open compressed log file: %v", err)
	}
	defer gzf.Close()

	gz := gzip.NewWriter(gzf)

	defer func() {
		if err != nil {
			os.Remove(dst)
			err = fmt.Errorf("failed to compress log file: %v", err)
		}
	}()

	if _, err := io.Copy(gz, f); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	if err := gzf.Close(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil {
		return err
	}

	return nil
}

// logInfo is a convenience struct to return the filename and its embedded
// timestamp.
type logInfo struct {
	timestamp time.Time
	Name      string
	Size      int64
}
