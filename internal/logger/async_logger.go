// ==============================================================================
// MasterDnsVPN — Fix: Async Logger
// Wraps Logger with a buffered channel so callers are never blocked waiting
// for console/file I/O. The background goroutine drains the channel serially,
// preserving log order while removing the mutex from the hot path.
// ==============================================================================
package logger

import (
	"sync"
)

const asyncLoggerBufferSize = 4096

type asyncEntry struct {
	level  int
	format string
	args   []any
}

// AsyncLogger wraps a Logger with a buffered channel.
// Drop policy: if the buffer is full, the entry is discarded (non-blocking).
type AsyncLogger struct {
	inner *Logger
	ch    chan asyncEntry
	once  sync.Once
	done  chan struct{}
	wg    sync.WaitGroup
}

// NewAsync wraps an existing Logger in an async adapter.
func NewAsync(inner *Logger) *AsyncLogger {
	a := &AsyncLogger{
		inner: inner,
		ch:    make(chan asyncEntry, asyncLoggerBufferSize),
		done:  make(chan struct{}),
	}
	a.wg.Add(1)
	go a.run()
	return a
}

func (a *AsyncLogger) run() {
	defer a.wg.Done()
	for {
		select {
		case e := <-a.ch:
			a.inner.logf(e.level, e.format, e.args...)
		case <-a.done:
			// Drain remaining entries before exit.
			for {
				select {
				case e := <-a.ch:
					a.inner.logf(e.level, e.format, e.args...)
				default:
					return
				}
			}
		}
	}
}

func (a *AsyncLogger) send(level int, format string, args []any) {
	select {
	case a.ch <- asyncEntry{level: level, format: format, args: args}:
	default:
		// Buffer full — drop to avoid blocking caller goroutine.
	}
}

func (a *AsyncLogger) Debugf(format string, args ...any) {
	if a.inner.Enabled(LevelDebug) {
		a.send(LevelDebug, format, args)
	}
}
func (a *AsyncLogger) Infof(format string, args ...any) {
	if a.inner.Enabled(LevelInfo) {
		a.send(LevelInfo, format, args)
	}
}
func (a *AsyncLogger) Warnf(format string, args ...any) {
	if a.inner.Enabled(LevelWarn) {
		a.send(LevelWarn, format, args)
	}
}
func (a *AsyncLogger) Errorf(format string, args ...any) {
	if a.inner.Enabled(LevelError) {
		a.send(LevelError, format, args)
	}
}
func (a *AsyncLogger) Enabled(level int) bool { return a.inner.Enabled(level) }

// Flush blocks until the channel is drained. Useful before process exit.
func (a *AsyncLogger) Flush() {
	a.once.Do(func() {
		close(a.done)
		a.wg.Wait()
	})
}
