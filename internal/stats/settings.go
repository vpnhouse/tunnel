package stats

import "time"

const (
	defaultFlushInterval  = 20 * time.Second
	defaultExportInterval = time.Minute
)

type Settings struct {
	FlushInterval  time.Duration `yaml:"flush_interval,omitempty"`
	ExportInterval time.Duration `yaml:"export_interval,omitempty"`
}

func (s *Settings) flushInterval() time.Duration {
	if s == nil || s.FlushInterval == 0 {
		return defaultFlushInterval
	}

	return s.FlushInterval
}

func (s *Settings) exportInterval() time.Duration {
	if s == nil || s.ExportInterval == 0 {
		return defaultExportInterval
	}

	return s.ExportInterval
}
