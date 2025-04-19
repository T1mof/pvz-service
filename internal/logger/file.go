package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileWriter struct {
	dir      string
	prefix   string
	maxSize  int64
	file     *os.File
	size     int64
	interval time.Duration
	lastTime time.Time
}

// NewFileWriter создает новый FileWriter
func NewFileWriter(dir, prefix string, maxSizeMB int, interval time.Duration) (*FileWriter, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию логов: %w", err)
	}

	w := &FileWriter{
		dir:      dir,
		prefix:   prefix,
		maxSize:  int64(maxSizeMB) * 1024 * 1024,
		interval: interval,
		lastTime: time.Now(),
	}

	if err := w.rotate(); err != nil {
		return nil, err
	}

	return w, nil
}

// Write реализует интерфейс io.Writer
func (w *FileWriter) Write(p []byte) (n int, err error) {
	now := time.Now()

	if w.interval > 0 && now.Sub(w.lastTime) >= w.interval {
		if err := w.rotate(); err != nil {
			return 0, err
		}
		w.lastTime = now
	}

	if w.maxSize > 0 && w.size+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// Close закрывает текущий файл
func (w *FileWriter) Close() error {
	if w.file == nil {
		return nil
	}
	return w.file.Close()
}

// rotate выполняет ротацию файла лога
func (w *FileWriter) rotate() error {
	if w.file != nil {
		w.file.Close()
	}

	timestamp := time.Now().Format("2006-01-02T15-04-05")
	filename := filepath.Join(w.dir, fmt.Sprintf("%s_%s.log", w.prefix, timestamp))

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл лога: %w", err)
	}

	w.file = f
	w.size = 0
	return nil
}
