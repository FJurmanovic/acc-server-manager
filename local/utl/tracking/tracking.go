package tracking

import (
	"bufio"
	"os"
	"time"
)

func TailLogFile(path string, callback func(string)) {
    file, _ := os.Open(path)
    defer file.Close()

    file.Seek(0, os.SEEK_END) // Start at end of file
    reader := bufio.NewReader(file)

    for {
        line, err := reader.ReadString('\n')
        if err == nil {
            callback(line)
        } else {
            time.Sleep(500 * time.Millisecond) // wait for new data
        }
    }
}