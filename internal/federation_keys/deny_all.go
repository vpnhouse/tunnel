package federation_keys

type DenyAllKeystore struct{}

func (DenyAllKeystore) Authorize(key string) (string, bool) {
	return "", false
}
