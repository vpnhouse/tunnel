package dnsbase

type LookupInterface interface {
	Lookup(domain string) (*Domain, error)
}

type WriterInterface interface {
	Write(domain string, cats Categories) error
	Close() error
}
