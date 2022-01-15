// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"bytes"

	"github.com/ava-labs/spacesvm/tdata"
	"github.com/ethereum/go-ethereum/common"
)

var _ UnsignedTransaction = &TransferTx{}

type TransferTx struct {
	*BaseTx `serialize:"true" json:"baseTx"`

	// To is the recipient of the [Units].
	To common.Address `serialize:"true" json:"to"`

	// Units are transferred to [To].
	Units uint64 `serialize:"true" json:"units"`
}

func (t *TransferTx) Execute(c *TransactionContext) error {
	// Must transfer to someone
	if bytes.Equal(t.To[:], zeroAddress[:]) {
		return ErrNonActionable
	}

	// This prevents someone from transferring to themselves.
	if bytes.Equal(t.To[:], c.Sender[:]) {
		return ErrNonActionable
	}
	if t.Units == 0 {
		return ErrNonActionable
	}
	if _, err := ModifyBalance(c.Database, c.Sender, false, t.Units); err != nil {
		return err
	}
	if _, err := ModifyBalance(c.Database, t.To, true, t.Units); err != nil {
		return err
	}
	return nil
}

func (t *TransferTx) Copy() UnsignedTransaction {
	to := make([]byte, common.AddressLength)
	copy(to, t.To[:])
	return &TransferTx{
		BaseTx: t.BaseTx.Copy(),
		To:     common.BytesToAddress(to),
		Units:  t.Units,
	}
}

func (t *TransferTx) TypedData() tdata.TypedData {
	return tdata.CreateTypedData(
		t.Magic, Transfer,
		[]tdata.Type{
			{Name: "blockID", Type: "string"},
			{Name: "price", Type: "uint64"},
			{Name: "to", Type: "address"},
			{Name: "units", Type: "uint64"},
		},
		tdata.TypedDataMessage{
			"blockID": t.BlockID.String(),
			"price":   t.Price,
			"to":      t.To,
			"units":   t.Units,
		},
	)
}
