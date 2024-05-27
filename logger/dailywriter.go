package logger

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type DailyWriter struct {
	file *os.File
	lock sync.Mutex

	folder      string
	prefix      string
	lastYearDay int
}

func NewDailyWriter(folder, prefix string) *DailyWriter {
	os.MkdirAll(folder, os.ModePerm)
	return &DailyWriter{folder: folder, prefix: prefix}
}

func (dw *DailyWriter) Write(p []byte) (n int, err error) {
	err = dw.checkNewDay()
	if err != nil {
		return
	}

	return dw.file.Write(p)
}

func (dw *DailyWriter) checkNewDay() error {
	dw.lock.Lock()
	defer dw.lock.Unlock()

	now := time.Now()
	yd := now.YearDay()

	if yd == dw.lastYearDay {
		return nil
	}

	ps := filepath.Join(dw.folder, dw.prefix+"-"+now.Format(time.DateOnly)+".log")

	logfile, err := os.OpenFile(ps, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	dw.lastYearDay = yd

	if dw.file != nil {
		dw.file.Close()
	}

	dw.file = logfile

	return nil
}
