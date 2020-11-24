package core

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/HPISTechnologies/mevm/geth/common"
)

func TestSequentialModeStateDB(t *testing.T) {
	a1 := common.BytesToAddress([]byte{1})
	a2 := common.BytesToAddress([]byte{2})
	a3 := common.BytesToAddress([]byte{3})
	s1 := string(a1.Bytes())
	s2 := string(a2.Bytes())
	// s3 := string(a3.Bytes())

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
				string(common.BytesToHash([]byte("key1")).Bytes()): string(common.BytesToHash([]byte("value1")).Bytes()),
			},
			s2: map[string]string{
				"key2": "value2",
			},
		},
	}

	state := NewStateDBInSequentialMode(mock, mock, nil)
	state.AddBalance(a1, new(big.Int).SetInt64(1))
	state.SubBalance(a2, new(big.Int).SetInt64(1))
	state.CreateAccount(a3)
	state.AddBalance(a3, new(big.Int).SetInt64(100))
	state.SetNonce(a1, 1)
	state.SetNonce(a2, 2)
	state.SetNonce(a3, 3)
	state.SetCode(a1, []byte{11})
	state.SetState(a2, common.BytesToHash([]byte("key1")), common.BytesToHash([]byte("value11")))
	state.SetState(a2, common.BytesToHash([]byte("key2")), common.BytesToHash([]byte("value22")))
	state.SetState(a3, common.BytesToHash([]byte("key1")), common.BytesToHash([]byte("value1")))
	// Commit.
	state.Prepare(common.Hash{}, common.Hash{}, 1)

	if state.GetBalance(a1).Cmp(new(big.Int).SetInt64(101)) != 0 ||
		state.GetNonce(a1) != 1 {
		t.Error("Checking acc1 failed")
		return
	}
	if bytes.Compare(state.GetCode(a1), []byte{11}) != 0 {
		t.Error("Checking code1 failed")
		return
	}
	if bytes.Compare(state.GetState(a1, common.BytesToHash([]byte("key1"))).Bytes(), common.BytesToHash([]byte("value1")).Bytes()) != 0 {
		t.Error("Checking state(a1, key1) failed")
		return
	}

	if state.GetBalance(a2).Cmp(new(big.Int).SetInt64(99)) != 0 ||
		state.GetNonce(a2) != 2 {
		t.Error("Checking acc2 failed")
		return
	}
	if bytes.Compare(state.GetCode(a2), []byte{2}) != 0 {
		t.Error("Checking code2 failed")
		return
	}
	if bytes.Compare(state.GetState(a2, common.BytesToHash([]byte("key1"))).Bytes(), common.BytesToHash([]byte("value11")).Bytes()) != 0 {
		t.Error("Checking state(a2, key1) failed")
		return
	}
	if bytes.Compare(state.GetState(a2, common.BytesToHash([]byte("key2"))).Bytes(), common.BytesToHash([]byte("value22")).Bytes()) != 0 {
		t.Error("Checking state(a2, key2) failed")
		return
	}

	if state.GetBalance(a3).Cmp(new(big.Int).SetInt64(100)) != 0 ||
		state.GetNonce(a3) != 3 {
		t.Error("Checking acc3 failed")
		return
	}
	if state.GetCode(a3) != nil {
		t.Error("Checking code3 failed")
		return
	}
	if bytes.Compare(state.GetState(a3, common.BytesToHash([]byte("key1"))).Bytes(), common.BytesToHash([]byte("value1")).Bytes()) != 0 {
		t.Error("Checking state(a3, key1) failed")
		return
	}
}
