package admin

import (
	"time"

	"github.com/vpnhouse/tunnel/internal/iprose"
	"github.com/vpnhouse/tunnel/internal/manager"
	"github.com/vpnhouse/tunnel/internal/storage"
)

type Service struct {
	// Manager to control over WG sessions stuff
	mgr *manager.Manager
	// Manager to control over IPRose sessions stuff
	ipr *iprose.Instance

	storage *storage.Storage
}

func New(mgr *manager.Manager, ipr *iprose.Instance, storage *storage.Storage) *Service {
	s := &Service{
		mgr:     mgr,
		ipr:     ipr,
		storage: storage,
	}

	go s.run()

	return s
}

func (s *Service) run() {
	s.storage.CleanupExpiredActionRules()
	for range time.Tick(time.Hour) {
		s.storage.CleanupExpiredActionRules()
	}
}
