package core

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/HPISTechnologies/mevm/geth/common"
)

type mockEthAccount struct {
	balance  *big.Int
	nonce    uint64
	codeHash []byte
}

func (mock *mockEthAccount) GetBalance() *big.Int {
	return mock.balance
}

func (mock *mockEthAccount) GetNonce() uint64 {
	return mock.nonce
}

func (mock *mockEthAccount) GetCodeHash() []byte {
	return mock.codeHash
}

type mockEthCache struct {
	accounts map[string]*mockEthAccount
	codes    map[string][]byte
	storages map[string]map[string]string
}

func (mock *mockEthCache) GetAccount(addr string) (Account, error) {
	if acc, ok := mock.accounts[addr]; ok {
		return acc, nil
	}
	return nil, nil
}

func (mock *mockEthCache) GetCode(addr string) ([]byte, error) {
	if code, ok := mock.codes[addr]; ok {
		return code, nil
	}
	return nil, nil
}

func (mock *mockEthCache) GetState(addr string, key []byte) []byte {
	if s, ok := mock.storages[addr]; ok {
		if v, ok := s[string(key)]; ok {
			return []byte(v)
		}
	}
	return nil
}

func TestDirtyCacheLocalCommit(t *testing.T) {
	a1 := common.BytesToAddress([]byte{1})
	a2 := common.BytesToAddress([]byte{2})
	a3 := common.BytesToAddress([]byte{3})
	s1 := string(a1.Bytes())
	s2 := string(a2.Bytes())
	s3 := string(a3.Bytes())

	mock := &mockEthCache{
		accounts: map[string]*mockEthAccount{
			s1: &mockEthAccount{
				balance: new(big.Int).SetInt64(100),
			},
			s2: &mockEthAccount{
				balance: new(big.Int).SetInt64(100),
			},
		},
		codes: map[string][]byte{
			s1: []byte{1},
			s2: []byte{2},
		},
		storages: map[string]map[string]string{
			s1: map[string]string{
				"key1": "value1",
			},
			s2: map[string]string{
				"key2": "value2",
			},
		},
	}
	dirtyCache := newDirtyCache(mock, mock)

	dirtyCache.localCommit(
		map[common.Address]struct{}{
			a3: struct{}{},
		},
		map[common.Address]*big.Int{
			a1: new(big.Int).SetInt64(1),
			a2: new(big.Int).SetInt64(-1),
			a3: new(big.Int).SetInt64(100),
		},
		map[common.Address]uint64{
			a1: 1,
			a2: 2,
			a3: 3,
		},
		map[common.Address][]byte{
			a1: []byte{11},
		},
		map[common.Address]map[common.Hash]common.Hash{
			a2: map[common.Hash]common.Hash{
				common.BytesToHash([]byte("key1")): common.BytesToHash([]byte("value11")),
				common.BytesToHash([]byte("key2")): common.BytesToHash([]byte("value22")),
			},
			a3: map[common.Hash]common.Hash{
				common.BytesToHash([]byte("key1")): common.BytesToHash([]byte("value1")),
			},
		})

	acc1, _ := dirtyCache.GetAccount(s1)
	if acc1.GetBalance().Cmp(new(big.Int).SetInt64(101)) != 0 ||
		acc1.GetNonce() != 1 {
		t.Error("Checking acc1 failed")
		return
	}
	code1, _ := dirtyCache.GetCode(s1)
	if bytes.Compare(code1, []byte{11}) != 0 {
		t.Error("Checking code1 failed")
		return
	}
	if bytes.Compare(dirtyCache.GetState(s1, []byte("key1")), []byte("value1")) != 0 {
		t.Error("Checking state(s1, key1) failed")
		return
	}

	acc2, _ := dirtyCache.GetAccount(s2)
	if acc2.GetBalance().Cmp(new(big.Int).SetInt64(99)) != 0 ||
		acc2.GetNonce() != 2 {
		t.Error("Checking acc2 failed")
		return
	}
	code2, _ := dirtyCache.GetCode(s2)
	if bytes.Compare(code2, []byte{2}) != 0 {
		t.Error("Checking code2 failed")
		return
	}
	if bytes.Compare(dirtyCache.GetState(s2, common.BytesToHash([]byte("key1")).Bytes()), common.BytesToHash([]byte("value11")).Bytes()) != 0 {
		t.Error("Checking state(s2, key1) failed")
		return
	}
	if bytes.Compare(dirtyCache.GetState(s2, common.BytesToHash([]byte("key2")).Bytes()), common.BytesToHash([]byte("value22")).Bytes()) != 0 {
		t.Error("Checking state(s2, key2) failed")
		return
	}

	acc3, _ := dirtyCache.GetAccount(s3)
	if acc3.GetBalance().Cmp(new(big.Int).SetInt64(100)) != 0 ||
		acc3.GetNonce() != 3 {
		t.Error("Checking acc3 failed")
		return
	}
	code3, _ := dirtyCache.GetCode(s3)
	if code3 != nil {
		t.Error("Checking code3 failed")
		return
	}
	if bytes.Compare(dirtyCache.GetState(s3, common.BytesToHash([]byte("key1")).Bytes()), common.BytesToHash([]byte("value1")).Bytes()) != 0 {
		t.Error("Checking state(s3, key1) failed")
		return
	}
}
