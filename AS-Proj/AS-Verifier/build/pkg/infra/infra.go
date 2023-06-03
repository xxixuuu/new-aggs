package infra

import (
	"fmt"
	"sync"

	"github.com/herumi/bls-go-binary/bls"
)

type PKMap struct {
	sync.Mutex
	dictionary map[string]*bls.PublicKey
}

//KEY = int64(AID):int64(SID), VALUE = bls.PublicKey
func (m *PKMap) Set(key string, value *bls.PublicKey) {
	m.Lock()
	m.dictionary[key] = value
	m.Unlock()
}

func (m *PKMap) Get(key string) *bls.PublicKey {
	m.Lock()
	value := m.dictionary[key]
	m.Unlock()
	return value
}

func (m *PKMap) GetSPKVec(key []string) []bls.PublicKey {
	var pkvec []bls.PublicKey = make([]bls.PublicKey, len(key))

	m.Lock()
	for i := 0; i < len(key); i++ {
		res := m.dictionary[key[i]]
		if res == nil {
			fmt.Println("In GetSPKVec null ptr, KEY=", key[i])
		}
		pkvec[i] = *m.dictionary[key[i]]
	}
	m.Unlock()
	return pkvec
}

func NewPKMap() *PKMap {
	return &PKMap{
		dictionary: make(map[string]*bls.PublicKey),
	}
}
