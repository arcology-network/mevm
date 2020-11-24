package state

import (
	"bytes"
	"math/big"

	"github.com/HPISTechnologies/mevm/geth/common"
	"github.com/HPISTechnologies/mevm/geth/core"
	"github.com/HPISTechnologies/mevm/geth/core/types"
	"github.com/HPISTechnologies/mevm/geth/crypto"
)

type account struct {
	balance  *big.Int
	nonce    uint64
	code     []byte
	codeHash common.Hash
	storage  map[common.Hash]common.Hash
	suicided bool
}

var emptyCodeHash = crypto.Keccak256(nil)

func newAccount() *account {
	return &account{
		balance:  big.NewInt(0),
		nonce:    0,
		code:     nil,
		codeHash: common.Hash{},
		storage:  make(map[common.Hash]common.Hash),
		suicided: false,
	}
}

func (acc *account) copy() *account {
	storage := make(map[common.Hash]common.Hash)
	for k, v := range acc.storage {
		storage[k] = v
	}

	return &account{
		balance:  new(big.Int).Set(acc.balance),
		nonce:    acc.nonce,
		code:     acc.code,
		codeHash: acc.codeHash,
		storage:  storage,
		suicided: acc.suicided,
	}
}

type StateDB struct {
	kapi    core.KernelAPI
	db      map[common.Address]*account
	refund  uint64
	thash   common.Hash
	logs    map[common.Hash][]*types.Log
	dirties map[common.Address]int64
}

func NewStateDB(kapi core.KernelAPI) *StateDB {
	return &StateDB{
		kapi:    kapi,
		db:      make(map[common.Address]*account),
		logs:    make(map[common.Hash][]*types.Log),
		dirties: make(map[common.Address]int64),
	}
}

// The following functions are used by EVM.

func (state *StateDB) CreateAccount(addr common.Address) {
	state.db[addr] = newAccount()
}

func (state *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	state.dirties[addr] -= amount.Int64()
}

func (state *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	state.dirties[addr] += amount.Int64()
}

func (state *StateDB) GetBalance(addr common.Address) *big.Int {
	if v, ok := state.dirties[addr]; ok {
		return new(big.Int).Add(state.db[addr].balance, new(big.Int).SetInt64(v))
	}
	return state.db[addr].balance
}

func (state *StateDB) GetBalanceNoRecord(addr common.Address) *big.Int {
	return state.GetBalance(addr)
}

func (state *StateDB) GetNonce(addr common.Address) uint64 {
	return state.db[addr].nonce
}

func (state *StateDB) SetNonce(addr common.Address, nonce uint64) {
	state.db[addr].nonce = nonce
}

func (state *StateDB) GetCodeHash(addr common.Address) common.Hash {
	if _, ok := state.db[addr]; !ok {
		state.db[addr] = newAccount()
	}
	return state.db[addr].codeHash
}

func (state *StateDB) GetCode(addr common.Address) []byte {
	return state.db[addr].code
}

func (state *StateDB) SetCode(addr common.Address, code []byte) {
	state.db[addr].code = code
	state.db[addr].codeHash = crypto.Keccak256Hash(code)
}

func (state *StateDB) GetCodeSize(addr common.Address) int {
	if state.kapi.IsKernelAPI(addr) {
		return 0xff
	}
	return len(state.db[addr].code)
}

func (state *StateDB) AddRefund(gas uint64) {
	state.refund += gas
}

func (state *StateDB) SubRefund(gas uint64) {
	state.refund -= gas
}

func (state *StateDB) GetRefund() uint64 {
	return state.refund
}

func (state *StateDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	return state.db[addr].storage[hash]
}

func (state *StateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	return state.db[addr].storage[hash]
}

func (state *StateDB) SetState(addr common.Address, key, value common.Hash) {
	state.db[addr].storage[key] = value
}

func (state *StateDB) Suicide(addr common.Address) bool {
	state.db[addr].suicided = true
	state.db[addr].balance = new(big.Int)
	return true
}

func (state *StateDB) HasSuicided(addr common.Address) bool {
	return state.db[addr].suicided
}

func (state *StateDB) Exist(addr common.Address) bool {
	_, exist := state.db[addr]
	return exist
}

func (state *StateDB) Empty(addr common.Address) bool {
	account := state.db[addr]
	return account == nil || (account.nonce == 0 && account.balance.Sign() == 0 && bytes.Equal(account.codeHash[:], emptyCodeHash))
}

func (state *StateDB) RevertToSnapshot(revid int) {

}

func (state *StateDB) Snapshot() int {
	return 0
}

func (state *StateDB) AddLog(log *types.Log) {
	state.logs[state.thash] = append(state.logs[state.thash], log)
}

func (state *StateDB) AddPreimage(hash common.Hash, preimage []byte) {

}

func (state *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) {

}

func (state *StateDB) Set(eac core.EthAccountCache, esc core.EthStorageCache) {

}

// The following functions are used by ExecutionUnit.

func (state *StateDB) Prepare(thash, bhash common.Hash, ti int) {
	state.thash = thash
	state.logs = make(map[common.Hash][]*types.Log)
}

func (state *StateDB) GetLogs(hash common.Hash) []*types.Log {
	return state.logs[hash]
}

// The following functions are used for test.

func (state *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	state.db[addr].balance = amount
}

// The following functions are used by Scheduler.

func (state *StateDB) Copy() core.StateDB {
	return &StateDB{
		db:      state.db,
		logs:    make(map[common.Hash][]*types.Log),
		dirties: make(map[common.Address]int64),
	}
}
