package openminder

import (
	"sync"
	"time"
)

func newErrorStore() *errorStore {
	return &errorStore{
		store: map[string]time.Time{},
		mu:    new(sync.Mutex),
	}
}

type errorStore struct {
	store map[string]time.Time
	mu    *sync.Mutex
}

func (es *errorStore) Add(err error) {
	if err == nil {
		return
	}

	es.mu.Lock()
	defer es.mu.Unlock()
	es.store[err.Error()] = time.Now()
}

func (es *errorStore) Map() map[string]string {
	data := map[string]string{}
	for msg, t := range es.store {
		data[msg] = time.Since(t).String()
	}
	return data
}

func (es *errorStore) Clean(err error) {
	for {
		// delete errors that haven't been reported for 10 secs
		for k, t := range es.store {
			if time.Since(t) > time.Minute*2 {
				delete(es.store, k)
			}
		}
	}
}
