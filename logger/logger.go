package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

const logo = `
	 ██████╗  ██████╗ ██╗	  ██████╗  █████╗ ██████╗ 
	██╔════╝ ██╔═══██╗██║	 ██╔═══██╗██╔══██╗██╔══██╗
	██║  ███╗██║   ██║██║	 ██║   ██║███████║██║  ██║
	██║   ██║██║   ██║██║	 ██║   ██║██╔══██║██║  ██║
	╚██████╔╝╚██████╔╝███████╗╚██████╔╝██║  ██║██████╔╝
	 ╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝ 
	════════════════════════════════════════════════════`

type Logger struct {
	Dateformat string
	input      io.Writer
}

func NewLogger(logDir string) (Logger, error) {
	logger := Logger{
		Dateformat: "2006-01-02 15:04:05",
		input:      nil,
	}
	fileName := logDir + "/log-" + time.Now().Format(logger.Dateformat) + ".log"
	file, err := os.Create(fileName)
	if err != nil {
		return logger, fmt.Errorf("error creating log file: %s", err)
	}
	logger.input = io.Writer(file)
	return logger, nil
}

func (logger *Logger) Log(logData string) error {
	logData = time.Now().Format(logger.Dateformat) + " " + logData + "\n"
	_, err := logger.input.Write([]byte(logData))
	if err != nil {
		return err
	}
	return nil
}

func (logger *Logger) LogLogo() error {
	logData := logo + "\n"
	_, err := logger.input.Write([]byte(logData))
	if err != nil {
		return err
	}
	return nil
}

func (logger *Logger) Close() error {
	logger.input = nil
	return nil
}
