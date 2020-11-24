package core

import (
	"fmt"
	"math/big"

	"github.com/HPISTechnologies/mevm/geth/common"
	"github.com/HPISTechnologies/mevm/geth/crypto"
)

type dirtyAccount struct {
	balance  *big.Int
	nonce    uint64
	code     []byte
	codeHash []byte
	storage  map[string]string
}

func newDirtyAccount() *dirtyAccount {
	return &dirtyAccount{
		balance: new(big.Int),
		storage: make(map[string]string),
	}
}

func newDirtyAccountFrom(acc Account) *dirtyAccount {
	return &dirtyAccount{
		balance:  acc.GetBalance(),
		nonce:    acc.GetNonce(),
		codeHash: acc.GetCodeHash(),
		storage:  make(map[string]string),
	}
}

func (da *dirtyAccount) GetBalance() *big.Int {
	return da.balance
}

func (da *dirtyAccount) GetNonce() uint64 {
	return da.nonce
}

func (da *dirtyAccount) GetCodeHash() []byte {
	return da.codeHash
}

type dirtyCache struct {
	accountCache EthAccountCache
	storageCache EthStorageCache
	dirties      map[string]*dirtyAccount
}

func newDirtyCache(accountCache EthAccountCache, storageCache EthStorageCache) *dirtyCache {
	return &dirtyCache{
		accountCache: accountCache,
		storageCache: storageCache,
		dirties:      make(map[string]*dirtyAccount),
	}
}

func (dc *dirtyCache) GetAccount(addr string) (Account, error) {
	if acc, ok := dc.dirties[addr]; ok {
		return acc, nil
	}
	return dc.accountCache.GetAccount(addr)
}

func (dc *dirtyCache) GetCode(addr string) ([]byte, error) {
	if acc, ok := dc.dirties[addr]; ok && acc.code != nil {
		return acc.code, nil
	}
	return dc.accountCache.GetCode(addr)
}

func (dc *dirtyCache) GetState(addr string, key []byte) []byte {
	if acc, ok := dc.dirties[addr]; ok {
		if value, ok := acc.storage[string(key)]; ok {
			return []byte(value)
		}
	}
	return dc.storageCache.GetState(addr, key)
}

func (dc *dirtyCache) localCommit(
	newlyCreated map[common.Address]struct{},
	balanceWrites map[common.Address]*big.Int,
	nonceWrites map[common.Address]uint64,
	codeWrites map[common.Address][]byte,
	storageWrites map[common.Address]map[common.Hash]common.Hash,
) {
	for addr := range newlyCreated {
		dc.dirties[string(addr.Bytes())] = newDirtyAccount()
	}

	for addr, amount := range balanceWrites {
		if acc, ok := dc.dirties[string(addr.Bytes())]; ok {
			acc.balance = new(big.Int).Add(acc.balance, amount)
		} else {
			acc, err := dc.accountCache.GetAccount(string(addr.Bytes()))
			if err != nil {
				panic(fmt.Sprintf("unexpected error: %v", err))
			}
			da := newDirtyAccountFrom(acc)
			da.balance = new(big.Int).Add(da.balance, amount)
			dc.dirties[string(addr.Bytes())] = da
		}
	}

	for addr, nonce := range nonceWrites {
		if acc, ok := dc.dirties[string(addr.Bytes())]; ok {
			acc.nonce = nonce
		} else {
			acc, err := dc.accountCache.GetAccount(string(addr.Bytes()))
			if err != nil {
				panic(fmt.Sprintf("unexpected error: %v", err))
			}
			da := newDirtyAccountFrom(acc)
			da.nonce = nonce
			dc.dirties[string(addr.Bytes())] = da
		}
	}

	for addr, code := range codeWrites {
		if acc, ok := dc.dirties[string(addr.Bytes())]; ok {
			acc.code = code
			acc.codeHash = crypto.Keccak256Hash(code).Bytes()
		} else {
			acc, err := dc.accountCache.GetAccount(string(addr.Bytes()))
			if err != nil {
				panic(fmt.Sprintf("unexpected error: %v", err))
			}
			da := newDirtyAccountFrom(acc)
			da.code = code
			da.codeHash = crypto.Keccak256Hash(code).Bytes()
			dc.dirties[string(addr.Bytes())] = da
		}
	}

	for addr, storage := range storageWrites {
		if acc, ok := dc.dirties[string(addr.Bytes())]; ok {
			for k, v := range storage {
				acc.storage[string(k.Bytes())] = string(v.Bytes())
			}
		} else {
			acc, err := dc.accountCache.GetAccount(string(addr.Bytes()))
			if err != nil {
				panic(fmt.Sprintf("unexpected error: %v", err))
			}
			da := newDirtyAccountFrom(acc)
			for k, v := range storage {
				da.storage[string(k.Bytes())] = string(v.Bytes())
			}
			dc.dirties[string(addr.Bytes())] = da
		}
	}
}
