package stats

import "time"

const (
	defaultFlushInterval  = time.Minute
	defaultExportInterval = 5 * time.Minute
)

type Settings struct {
	FlushInterval  time.Duration `yaml:"flush_interval,omitempty"`
	ExportInterval time.Duration `yaml:"export_interval,omitempty"`
}

func (s *Settings) flushInterval() time.Duration {
	if s.FlushInterval == 0 {
		return defaultFlushInterval
	}

	return s.FlushInterval
}

func (s *Settings) exportInterval() time.Duration {
	if s.ExportInterval == 0 {
		return defaultExportInterval
	}

	return s.ExportInterval
}
