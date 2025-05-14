package stats

type ExtraStats struct {
	PeersTotal  int
	PeersActive int
}

type ExtraStatsCb func() ExtraStats

type Stats struct {
	ExtraStats
	UpstreamBytes   int64
	DownstreamBytes int64
	UpstreamSpeed   int64
	DownstreamSpeed int64
}

func (s *Stats) Add(v *Stats) {
	s.PeersActive += v.PeersActive
	s.PeersTotal += v.PeersTotal
	s.UpstreamBytes += v.UpstreamBytes
	s.UpstreamSpeed += v.UpstreamSpeed
	s.DownstreamBytes += v.DownstreamBytes
	s.DownstreamSpeed += v.DownstreamSpeed
}
