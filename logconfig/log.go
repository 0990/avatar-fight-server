package logconfig

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"io"
)

func InitLogrus(name string, maxMB int) {
	formatter := &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.AddHook(NewDefaultHook(name, maxMB))
}

type DefaultHook struct {
	writers   map[logrus.Level]io.Writer
	errWriter io.Writer
	fmt       logrus.Formatter
}

func NewDefaultHook(name string, maxSize int) *DefaultHook {
	formatter := &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: false,
	}

	writers := make(map[logrus.Level]io.Writer)
	for _, level := range logrus.AllLevels {
		writers[level] = &lumberjack.Logger{
			Filename:   fmt.Sprintf("%s_%s.log", name, level.String()),
			MaxSize:    maxSize,
			MaxAge:     100,
			MaxBackups: 100,
			LocalTime:  true,
			Compress:   false,
		}
	}

	return &DefaultHook{
		writers: writers,
		fmt:     formatter,
	}
}

func (p *DefaultHook) Fire(entry *logrus.Entry) error {
	data, err := p.fmt.Format(entry)
	if err != nil {
		return err
	}
	_, err = p.writers[entry.Level].Write(data)
	return err
}

func (p *DefaultHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
