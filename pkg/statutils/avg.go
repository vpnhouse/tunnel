package statutils

type AvgValue struct {
	values []int64
	i      int
}

func NewAvgValue(size int) *AvgValue {
	if size <= 0 {
		size = 10
	}
	return &AvgValue{
		values: make([]int64, 0, size),
		i:      -1,
	}
}

func (s *AvgValue) Push(val int64) int64 {
	if s.i+1 == cap(s.values) {
		s.i = -1
	}
	s.i++
	if s.i == len(s.values) {
		s.values = append(s.values, val)
	} else {
		s.values[s.i] = val
	}
	return s.Avg()
}

func (s *AvgValue) Avg() int64 {
	if len(s.values) == 0 {
		return 0
	}
	var sum int64
	for _, val := range s.values {
		sum += val
	}
	return sum / int64(len(s.values))
}
