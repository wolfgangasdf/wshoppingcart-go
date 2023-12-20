package jeffdisk

// this is just a copy of https://github.com/abraithwaite/jeff/blob/main/memory/memory.go
// with persist to disk and load...

import (
	"context"
	"encoding/gob"
	"os"
	"sync"
	"time"
)

type item struct {
	Value []byte // WL exported these fields to make gob encode it.
	Exp   time.Time
}

var now = func() time.Time {
	return time.Now()
}

// Memory satisfies the jeff.Storage interface
type Memory struct {
	sessions map[string]item
	rw       sync.RWMutex
	filename string
}

// New initializes a new in-memory Storage for jeff
func New(filename string) *Memory {
	s := make(map[string]item)
	// WL load from file if exists
	file, err := os.Open(filename)
	if err == nil {
		dec := gob.NewDecoder(file)
		err := dec.Decode(&s)
		if err != nil {
			panic(err)
		}
	}
	defer file.Close()
	return &Memory{
		sessions: s,
		filename: filename,
	}
}

// WL save sessions to file
func (m *Memory) saveToFile() {
	// WL save sessions to file
	file, err := os.Create(m.filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	gob.Register(m.sessions)
	err = enc.Encode(m.sessions)
	if err != nil {
		panic(err)
	}

}

// Store satisfies the jeff.Store.Store method
func (m *Memory) Store(_ context.Context, key, value []byte, exp time.Time) error {
	m.rw.Lock()
	m.sessions[string(key)] = item{
		Value: value,
		Exp:   exp,
	}
	m.saveToFile()
	m.rw.Unlock()
	return nil
}

// Fetch satisfies the jeff.Store.Fetch method
func (m *Memory) Fetch(_ context.Context, key []byte) ([]byte, error) {
	m.rw.RLock()
	v, ok := m.sessions[string(key)]
	m.rw.RUnlock()
	if !ok || v.Exp.Before(time.Now()) {
		return nil, nil
	}
	return v.Value, nil
}

// Delete satisfies the jeff.Store.Delete method
func (m *Memory) Delete(_ context.Context, key []byte) error {
	m.rw.Lock()
	delete(m.sessions, string(key))
	m.saveToFile()
	m.rw.Unlock()
	return nil
}
