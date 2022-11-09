package xlimits

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/muesli/cache2go"
	"go.uber.org/zap"
)

type Recent struct {
	recent *cache2go.CacheTable
	max    int
	period time.Duration
	lock   sync.Mutex
}

func NewRecent(max int, period time.Duration) *Recent {
	return &Recent{
		recent: cache2go.Cache(uuid.NewString()),
		max:    max,
		period: period,
	}
}

func (r *Recent) Hit(ip string) bool {
	if ip == "" {
		return false
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	var cntr *int
	cached, err := r.recent.Value(ip)

	if err == nil {
		cntr = cached.Data().(*int)
		if cntr != nil {
			*cntr += 1
			return *cntr > r.max
		} else {
			zap.L().Error("Invalid nil recent pointer")
		}
	}

	var cntr_value int = 1
	cntr = &cntr_value

	r.recent.Add(ip, r.period, cntr)
	return cntr_value > r.max
}

func (r *Recent) Undo(ip string) int {
	if ip == "" {
		return 0
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	var cntr *int
	cached, err := r.recent.Value(ip)
	if err != nil {
		return 0
	}

	cntr = cached.Data().(*int)
	if cntr == nil {
		return 0
	}

	*cntr -= 1
	if *cntr == 0 {
		r.recent.Delete(ip)
	}

	return *cntr
}
