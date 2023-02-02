package eventlog

type Offset struct {
	Node   string
	LogID  string
	Offset int64
}
