package types

import (
	"math/big"

	"github.com/HPISTechnologies/mevm/geth/common"
)

type Reads struct {
	BalanceReads    map[common.Address]*big.Int
	EthStorageReads []common.Address
}

type Writes struct {
	NewAccounts      []common.Address
	BalanceWrites    map[common.Address]*big.Int
	BalanceOrigin    map[common.Address]*big.Int
	NonceWrites      map[common.Address]uint64
	CodeWrites       map[common.Address][]byte
	EthStorageWrites map[common.Address]map[common.Hash]common.Hash
}

type EuResult struct {
	H       common.Hash
	R       *Reads
	W       *Writes
	Status  uint64
	GasUsed uint64
}
