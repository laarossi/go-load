package utils

import (
	"fmt"
	"goload/types"
	"os"
	"sync"
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
	input      *SafeFileWriter
}

func NewLogger(logDir string) (*Logger, error) {
	logger := Logger{
		Dateformat: "2006-01-02-15:04:05",
		input:      nil,
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating directory: %s", err)
	}
	fileName := logDir + "/log-" + time.Now().Format(logger.Dateformat) + ".log"
	file, err := os.Create(fileName)
	if err != nil {
		return &logger, fmt.Errorf("error creating log file: %s", err)
	}
	logger.input = &SafeFileWriter{
		File: file,
		Mu:   sync.Mutex{},
	}
	logger.logLogo()
	return &logger, nil
}

func (logger *Logger) Log(logData string) error {
	logData = time.Now().Format(logger.Dateformat) + " " + logData + "\n"
	_, err := logger.input.Write(logData)
	if err != nil {
		return err
	}
	return nil
}

func (logger *Logger) LogResponse(response types.HTTPResponse) error {

	networkStats := ""
	if response.NetworkMetric != nil {
		networkStats = fmt.Sprintf("sent=%d bytes | recv=%d bytes",
			response.NetworkMetric.BytesSent,
			response.NetworkMetric.BytesRecv)
	}

	logData := fmt.Sprintf("%s [executor-%06d] status=%d resp_time=%06dms | %s | %s\n",
		time.Now().Format(logger.Dateformat),
		GetGoroutineID(),
		response.StatusCode,
		response.RequestMetric.Duration.Milliseconds(),
		networkStats,
		response.Body)

	_, err := logger.input.Write(logData)
	if err != nil {
		return err
	}
	return nil
}

func (logger *Logger) logLogo() error {
	logData := logo + "\n"
	_, err := logger.input.Write(logData)
	if err != nil {
		return err
	}
	return nil
}

func (logger *Logger) LogSeparator() error {
	_, err := logger.input.Write("================================================================\n")
	if err != nil {
		return err
	}
	return nil
}

func (logger *Logger) LogWithoutDate(logData string) error {
	_, err := logger.input.Write(logData + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (logger *Logger) Close() error {
	logger.input = nil
	return nil
}
