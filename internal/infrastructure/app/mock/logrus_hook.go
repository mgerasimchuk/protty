package mock

import (
	"github.com/sirupsen/logrus"
)

type LogrusHook struct {
	levels    []logrus.Level
	entryChan chan *logrus.Entry
}

func NewLogrusHook(levels []logrus.Level, chanBufferSize int) *LogrusHook {
	return &LogrusHook{levels: levels, entryChan: make(chan *logrus.Entry, chanBufferSize)}
}

func (h *LogrusHook) Levels() []logrus.Level {
	return h.levels
}

func (h *LogrusHook) Fire(entry *logrus.Entry) error {
	h.entryChan <- entry
	return nil
}

func (h *LogrusHook) EntryChan() chan *logrus.Entry {
	return h.entryChan
}
