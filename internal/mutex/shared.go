package mutex

import "sync"

type (
	Exclusion interface {
		Exit()
	}

	Zone struct {
		sync.RWMutex
	}

	sharedZone struct {
		zone *Zone
	}

	exclusiveZone struct {
		zone *Zone
	}
)

func Run(ex Exclusion, fn func()) {
	defer ex.Exit()
	fn()
}

func RunErr(ex Exclusion, fn func() error) error {
	defer ex.Exit()
	return fn()
}

func (z *Zone) Shared() Exclusion {
	z.RLock()
	return sharedZone{zone: z}
}

func (z *Zone) Exclusive() Exclusion {
	z.Lock()
	return exclusiveZone{zone: z}
}

func (ez exclusiveZone) Exit() {
	ez.zone.Unlock()
}

func (s sharedZone) Exit() {
	s.zone.RUnlock()
}
