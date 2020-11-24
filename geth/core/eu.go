package core

import (
	"math/big"

	"github.com/HPISTechnologies/mevm/geth/common"
	"github.com/HPISTechnologies/mevm/geth/core/types"
	"github.com/HPISTechnologies/mevm/geth/core/vm"
	"github.com/HPISTechnologies/mevm/geth/crypto"
)

type EU struct {
	evm   *vm.EVM
	state StateDB
	kapi  KernelAPI
}

func NewEU(euID uint16, state StateDB, kapi KernelAPI, cfg *Config) *EU {
	return &EU{
		evm:   vm.NewEVM(NewEVMContext(cfg), state, cfg.ChainConfig, *cfg.VMConfig, kapi),
		state: state,
		kapi:  kapi,
	}
}

func (eu *EU) SetApc(eac EthAccountCache, esc EthStorageCache) {
	eu.state.Set(eac, esc)
	// eu.kapi.SetSnapshot(snapshot)
}

func (eu *EU) Run(hash common.Hash, msg *types.Message, coinbase common.Address) (*types.EuResult, *types.Receipt) {
	eu.state.Prepare(hash, common.Hash{}, 0)
	eu.kapi.Prepare(hash)

	eu.evm.Context.Coinbase = coinbase
	eu.evm.Context = ResetEVMContext(eu.evm.Context, *msg)

	_, gas, failed, _ := ApplyMessage(eu.evm, *msg)

	var result *types.EuResult = nil
	if !failed {
		// rs, ws := eu.kapi.Collect()
		storageReads := make([]common.Address, 0, len(eu.state.(*ethState).storageReads))
		for acc := range eu.state.(*ethState).storageReads {
			storageReads = append(storageReads, acc)
		}
		reads := &types.Reads{
			// ClibReads:       rs,
			BalanceReads:    eu.state.(*ethState).balanceReads,
			EthStorageReads: storageReads,
		}
		newAccounts := make([]common.Address, 0, len(eu.state.(*ethState).newlyCreated))
		for acc := range eu.state.(*ethState).newlyCreated {
			newAccounts = append(newAccounts, acc)
		}
		writes := &types.Writes{
			// ClibWrites:       ws,
			NewAccounts:      newAccounts,
			BalanceWrites:    eu.state.(*ethState).balanceWrites,
			NonceWrites:      eu.state.(*ethState).nonceWrites,
			CodeWrites:       eu.state.(*ethState).codeWrites,
			EthStorageWrites: eu.state.(*ethState).storageWrites,
		}
		result = &types.EuResult{
			R: reads,
			W: writes,
		}
	} else {
		balanceWrites := make(map[common.Address]*big.Int)
		balanceWrites[eu.evm.Coinbase] = new(big.Int).Mul(new(big.Int).SetUint64(gas), msg.GasPrice())
		balanceWrites[msg.From()] = new(big.Int).Neg(balanceWrites[eu.evm.Coinbase])
		writes := &types.Writes{
			BalanceWrites: balanceWrites,
		}
		result = &types.EuResult{
			W: writes,
		}
	}

	balanceOrigin := make(map[common.Address]*big.Int)
	for addr := range result.W.BalanceWrites {
		balanceOrigin[addr] = eu.state.(*ethState).GetBalanceCommitted(addr)
	}
	result.W.BalanceOrigin = balanceOrigin

	receipt := types.NewReceipt(nil, failed, gas)
	receipt.TxHash = hash
	receipt.GasUsed = gas
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(eu.evm.Context.Origin, msg.Nonce())
	}
	receipt.Logs = eu.state.GetLogs(hash)
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	result.H = hash
	result.Status = receipt.Status
	result.GasUsed = receipt.GasUsed

	return result, receipt
}
