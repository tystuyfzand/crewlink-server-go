package server

import "sync"

type PlayerIdMap struct {
	m map[string]uint64
	l *sync.RWMutex
}

func (p *PlayerIdMap) Set(cid string, id uint64) {
	p.l.Lock()
	defer p.l.Unlock()

	p.m[cid] = id
}

func (p *PlayerIdMap) Get(cid string) (uint64, bool) {
	p.l.RLock()
	defer p.l.RUnlock()

	id, exists := p.m[cid]

	return id, exists
}

func (p *PlayerIdMap) Remove(cid string) {
	p.l.Lock()
	defer p.l.Unlock()

	delete(p.m, cid)
}

func (p *PlayerIdMap) MapOf(cids []string) map[string]uint64 {
	p.l.RLock()
	defer p.l.RUnlock()

	out := make(map[string]uint64)

	for _, cid := range cids {
		if id, exists := p.m[cid]; exists {
			out[cid] = id
		}
	}

	return out
}
