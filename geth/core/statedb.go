package core

import (
	"math/big"

	"github.com/HPISTechnologies/mevm/geth/common"
	"github.com/HPISTechnologies/mevm/geth/core/types"
	"github.com/HPISTechnologies/mevm/geth/crypto"
)

type ethState struct {
	accountCache EthAccountCache
	storageCache EthStorageCache
	kapi         KernelAPI
	refund       uint64
	thash        common.Hash
	logs         map[common.Hash][]*types.Log
	// Reads
	balanceReads map[common.Address]*big.Int
	storageReads map[common.Address]struct{}
	// Writes
	newlyCreated  map[common.Address]struct{}
	balanceWrites map[common.Address]*big.Int
	nonceWrites   map[common.Address]uint64
	codeWrites    map[common.Address][]byte
	storageWrites map[common.Address]map[common.Hash]common.Hash
	seqMode       bool
}

// NewStateDB creates an instance of ethState and returns it as an StateDB.
func NewStateDB(eac EthAccountCache, esc EthStorageCache, kapi KernelAPI) StateDB {
	return &ethState{
		accountCache:  eac,
		storageCache:  esc,
		kapi:          kapi,
		logs:          make(map[common.Hash][]*types.Log),
		balanceReads:  make(map[common.Address]*big.Int),
		storageReads:  make(map[common.Address]struct{}),
		newlyCreated:  make(map[common.Address]struct{}),
		balanceWrites: make(map[common.Address]*big.Int),
		nonceWrites:   make(map[common.Address]uint64),
		codeWrites:    make(map[common.Address][]byte),
		storageWrites: make(map[common.Address]map[common.Hash]common.Hash),
	}
}

func NewStateDBInSequentialMode(eac EthAccountCache, esc EthStorageCache, kapi KernelAPI) StateDB {
	cache := newDirtyCache(eac, esc)
	db := NewStateDB(cache, cache, kapi)
	db.(*ethState).seqMode = true
	return db
}

func (es *ethState) Set(eac EthAccountCache, esc EthStorageCache) {
	if es.seqMode {
		cache := newDirtyCache(eac, esc)
		es.accountCache = cache
		es.storageCache = cache
	} else {
		es.accountCache = eac
		es.storageCache = esc
	}
}

// CreateAccount creates an empty account object.
// It will not check if the addr exist.
func (es *ethState) CreateAccount(addr common.Address) {
	es.newlyCreated[addr] = struct{}{}
}

func (es *ethState) SubBalance(addr common.Address, amount *big.Int) {
	if v, ok := es.balanceWrites[addr]; ok {
		es.balanceWrites[addr] = new(big.Int).Sub(v, amount)
	} else {
		es.balanceWrites[addr] = new(big.Int).Neg(amount)
	}
}

func (es *ethState) AddBalance(addr common.Address, amount *big.Int) {
	if v, ok := es.balanceWrites[addr]; ok {
		es.balanceWrites[addr] = new(big.Int).Add(v, amount)
	} else {
		es.balanceWrites[addr] = amount
	}
}

func (es *ethState) GetBalance(addr common.Address) *big.Int {
	amount := es.GetBalanceNoRecord(addr)
	es.balanceReads[addr] = new(big.Int).Set(amount)
	return amount
}

func (es *ethState) GetBalanceNoRecord(addr common.Address) *big.Int {
	amount := es.GetBalanceCommitted(addr)
	if v, ok := es.balanceWrites[addr]; ok {
		return new(big.Int).Add(amount, v)
	}
	return amount
}

func (es *ethState) GetBalanceCommitted(addr common.Address) *big.Int {
	var amount *big.Int
	if acc, _ := es.accountCache.GetAccount(string(addr.Bytes())); acc == nil {
		amount = new(big.Int).SetInt64(0)
	} else {
		amount = acc.GetBalance()
	}
	return amount
}

// SetBalance is for test only.
func (es *ethState) SetBalance(addr common.Address, amount *big.Int) {
	es.balanceWrites[addr] = amount
}

func (es *ethState) GetNonce(addr common.Address) uint64 {
	if v, ok := es.nonceWrites[addr]; ok {
		return v
	}

	var nonce uint64
	if acc, _ := es.accountCache.GetAccount(string(addr.Bytes())); acc == nil {
		nonce = 0
	} else {
		nonce = acc.GetNonce()
	}
	return nonce
}

func (es *ethState) SetNonce(addr common.Address, nonce uint64) {
	es.nonceWrites[addr] = nonce
}

func (es *ethState) GetCodeHash(addr common.Address) common.Hash {
	if v, ok := es.codeWrites[addr]; ok {
		return crypto.Keccak256Hash(v)
	}

	var hash common.Hash
	if acc, _ := es.accountCache.GetAccount(string(addr.Bytes())); acc == nil {
		hash = common.Hash{}
	} else {
		hash = common.BytesToHash(acc.GetCodeHash())
	}
	return hash
}

func (es *ethState) GetCode(addr common.Address) []byte {
	if v, ok := es.codeWrites[addr]; ok {
		return v
	}

	if code, _ := es.accountCache.GetCode(string(addr.Bytes())); code != nil {
		return code
	}
	return nil
}

func (es *ethState) SetCode(addr common.Address, code []byte) {
	es.codeWrites[addr] = code
}

func (es *ethState) GetCodeSize(addr common.Address) int {
	if es.kapi.IsKernelAPI(addr) {
		// FIXME!
		return 0xff
	}

	return len(es.GetCode(addr))
}

func (es *ethState) AddRefund(amount uint64) {
	es.refund += amount
}

func (es *ethState) SubRefund(amount uint64) {
	es.refund -= amount
}

func (es *ethState) GetRefund() uint64 {
	return es.refund
}

func (es *ethState) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	value := es.storageCache.GetState(string(addr.Bytes()), key.Bytes())
	es.storageReads[addr] = struct{}{}
	return common.BytesToHash(value)
}

func (es *ethState) GetState(addr common.Address, key common.Hash) common.Hash {
	if storage, ok := es.storageWrites[addr]; ok {
		if v, ok := storage[key]; ok {
			return v
		}
	}

	return es.GetCommittedState(addr, key)
}

func (es *ethState) SetState(addr common.Address, key, value common.Hash) {
	if _, ok := es.storageWrites[addr]; !ok {
		es.storageWrites[addr] = make(map[common.Hash]common.Hash)
	}
	es.storageWrites[addr][key] = value
}

func (es *ethState) Suicide(addr common.Address) bool {
	return true
}

func (es *ethState) HasSuicided(addr common.Address) bool {
	return false
}

func (es *ethState) Exist(addr common.Address) bool {
	if _, ok := es.newlyCreated[addr]; ok {
		return true
	}

	acc, _ := es.accountCache.GetAccount(string(addr.Bytes()))
	return acc != nil
}

func (es *ethState) Empty(addr common.Address) bool {
	return es.GetBalance(addr).Cmp(new(big.Int).SetInt64(0)) == 0 &&
		es.GetNonce(addr) == 0 &&
		es.GetCode(addr) == nil
}

func (es *ethState) RevertToSnapshot(id int) {
	es.newlyCreated = make(map[common.Address]struct{})
	es.balanceWrites = make(map[common.Address]*big.Int)
	es.nonceWrites = make(map[common.Address]uint64)
	es.codeWrites = make(map[common.Address][]byte)
	es.storageWrites = make(map[common.Address]map[common.Hash]common.Hash)
}

func (es *ethState) Snapshot() int {
	return 0
}

func (es *ethState) AddLog(log *types.Log) {
	es.logs[es.thash] = append(es.logs[es.thash], log)
}

func (es *ethState) AddPreimage(hash common.Hash, preimage []byte) {

}

func (es *ethState) ForEachStorage(addr common.Address, f func(common.Hash, common.Hash) bool) {

}

func (es *ethState) Prepare(thash, bhash common.Hash, ti int) {
	if es.seqMode {
		es.accountCache.(*dirtyCache).localCommit(
			es.newlyCreated,
			es.balanceWrites,
			es.nonceWrites,
			es.codeWrites,
			es.storageWrites,
		)
	}
	es.thash = thash
	es.refund = 0
	es.balanceReads = make(map[common.Address]*big.Int)
	es.storageReads = make(map[common.Address]struct{})
	es.newlyCreated = make(map[common.Address]struct{})
	es.balanceWrites = make(map[common.Address]*big.Int)
	es.nonceWrites = make(map[common.Address]uint64)
	es.codeWrites = make(map[common.Address][]byte)
	es.storageWrites = make(map[common.Address]map[common.Hash]common.Hash)
	es.logs = make(map[common.Hash][]*types.Log)
}

func (es *ethState) GetLogs(hash common.Hash) []*types.Log {
	return es.logs[hash]
}

func (es *ethState) Copy() StateDB {
	return &ethState{
		accountCache:  es.accountCache,
		storageCache:  es.storageCache,
		logs:          make(map[common.Hash][]*types.Log),
		balanceReads:  make(map[common.Address]*big.Int),
		storageReads:  make(map[common.Address]struct{}),
		newlyCreated:  make(map[common.Address]struct{}),
		balanceWrites: make(map[common.Address]*big.Int),
		nonceWrites:   make(map[common.Address]uint64),
		codeWrites:    make(map[common.Address][]byte),
		storageWrites: make(map[common.Address]map[common.Hash]common.Hash),
	}
}
