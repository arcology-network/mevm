package types

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/HPISTechnologies/mevm/geth/common"
	"github.com/HPISTechnologies/mevm/geth/rlp"
)

var (
	nilAddress = common.Address{}
)

type Messager struct {
	Txhash common.Hash
	Msg    *Message
}

type MessagerRlp struct {
	Txhash common.Hash
	Msg    *MessageRlp
}

func ParseByte(data []byte) (*Messager, error) {
	//var ai *Messager
	var mrlp *MessagerRlp
	err := rlp.DecodeBytes(data, &mrlp)
	if err != nil {
		fmt.Printf("Message.FromBytes err=%v\n", err)
		return nil, err
	}

	to := mrlp.Msg.To
	if *to == nilAddress {
		to = nil
	}

	mi := &Messager{
		Txhash: mrlp.Txhash,
		Msg: &Message{
			to:         to,
			from:       mrlp.Msg.From,
			nonce:      mrlp.Msg.Nonce,
			amount:     mrlp.Msg.Amount,
			gasLimit:   mrlp.Msg.GasLimit,
			gasPrice:   mrlp.Msg.GasPrice,
			data:       mrlp.Msg.Data,
			checkNonce: mrlp.Msg.CheckNonce,
		},
	}
	return mi, nil
}
func (mi *Messager) ToByte() ([]byte, error) {

	to := mi.Msg.to
	if mi.Msg.to == nil {
		to = &nilAddress
	}

	mrlp := MessagerRlp{
		Txhash: mi.Txhash,
		Msg: &MessageRlp{
			To:         to,
			From:       mi.Msg.from,
			Nonce:      mi.Msg.nonce,
			Amount:     mi.Msg.amount,
			GasLimit:   mi.Msg.gasLimit,
			GasPrice:   mi.Msg.gasPrice,
			Data:       mi.Msg.data,
			CheckNonce: mi.Msg.checkNonce,
		},
	}
	return rlp.EncodeToBytes(mrlp)
}

type MessageRlp struct {
	To         *common.Address
	From       common.Address
	Nonce      uint64
	Amount     *big.Int
	GasLimit   uint64
	GasPrice   *big.Int
	Data       []byte
	CheckNonce bool
}

type Messagers struct {
	Msgs *[]*Messager
}

func (ms *Messagers) GetItemData() ([]byte, error) {
	return nil, nil
}

func (ms *Messagers) SetItemData(data []byte) error {
	return nil
}
func (ms *Messagers) GetItemSize() []int {
	sizes := make([]int, 1)
	sizes[0] = 130
	return sizes
}
func (ms *Messagers) GetListMeta() []int {

	sizes := make([]int, 1)
	sizes[0] = len(*ms.Msgs)
	return sizes

}

func (ms *Messagers) SetListMeta(idx int, size int) error {

	switch idx {
	case 0:
		lst := make([]*Messager, size)
		ms.Msgs = &lst
	}
	return nil
}

func (ms *Messagers) GetListItem(varIdx byte, idx int) ([]byte, error) {

	switch varIdx {
	case byte(0):
		if idx >= len(*ms.Msgs) {
			return nil, errors.New("idx is out of index")
		}
		mi := (*ms.Msgs)[idx]

		/*
			mrlp := MessagerRlp{
				Txhash: mi.Txhash,
				Msg: &MessageRlp{
					//To:         mi.Msg.to,
					From:       mi.Msg.from,
					Nonce:      mi.Msg.nonce,
					Amount:     mi.Msg.amount,
					GasLimit:   mi.Msg.gasLimit,
					GasPrice:   mi.Msg.gasPrice,
					Data:       mi.Msg.data,
					CheckNonce: mi.Msg.checkNonce,
				},
			}
			fmt.Printf("GetListItem----------------%v\n", mrlp)
			return rlp.EncodeToBytes(mrlp)
		*/
		return mi.ToByte()

	}

	return nil, errors.New("varIdx is out of index")

}
func (ms *Messagers) SetListItem(varIdx byte, idx int, data []byte) error {

	switch varIdx {
	case byte(0):
		if idx >= len(*ms.Msgs) {
			return errors.New("idx is out of index")
		}

		/*
			//var ai *Messager
			var mrlp *MessagerRlp
			err := rlp.DecodeBytes(data, &mrlp)
			if err != nil {
				fmt.Printf("Message.FromBytes err=%v\n", err)
				return err
			}

			mi := Messager{
				Txhash: mrlp.Txhash,
				Msg: &Message{
					//to:         mrlp.Msg.To,
					from:       mrlp.Msg.From,
					nonce:      mrlp.Msg.Nonce,
					amount:     mrlp.Msg.Amount,
					gasLimit:   mrlp.Msg.GasLimit,
					gasPrice:   mrlp.Msg.GasPrice,
					data:       mrlp.Msg.Data,
					checkNonce: mrlp.Msg.CheckNonce,
				},
			}
		*/
		mi, err := ParseByte(data)
		if err != nil {
			return err
		}
		(*ms.Msgs)[idx] = mi
		return nil
	}

	return errors.New("varIdx is out of index")

}
