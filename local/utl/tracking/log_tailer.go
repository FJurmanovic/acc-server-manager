package tracking

import (
	"bufio"
	"os"
	"time"
)

type LogTailer struct {
	filePath   string
	handleLine func(string)
	stopChan   chan struct{}
	isRunning  bool
	tracker    *PositionTracker
}

func NewLogTailer(filePath string, handleLine func(string)) *LogTailer {
	return &LogTailer{
		filePath:   filePath,
		handleLine: handleLine,
		stopChan:   make(chan struct{}),
		tracker:    NewPositionTracker(filePath),
	}
}

func (t *LogTailer) Start() {
	if t.isRunning {
		return
	}
	t.isRunning = true

	go func() {
		pos, err := t.tracker.LoadPosition()
		if err != nil {
			pos = &LogPosition{}
		}
		lastSize := pos.LastPosition

		for {
			select {
			case <-t.stopChan:
				t.isRunning = false
				return
			default:
				if file, err := os.Open(t.filePath); err == nil {
					stat, err := file.Stat()
					if err != nil {
						file.Close()
						time.Sleep(time.Second)
						continue
					}

					if stat.Size() < lastSize {
						lastSize = 0
					}

					if lastSize > 0 {
						file.Seek(lastSize, 0)
					}

					scanner := bufio.NewScanner(file)
					for scanner.Scan() {
						line := scanner.Text()
						t.handleLine(line)
						lastSize, _ = file.Seek(0, 1)

						t.tracker.SavePosition(&LogPosition{
							LastPosition: lastSize,
							LastRead:     line,
						})
					}

					file.Close()
				}

				time.Sleep(time.Second)
			}
		}
	}()
}

func (t *LogTailer) Stop() {
	if !t.isRunning {
		return
	}
	close(t.stopChan)
}
