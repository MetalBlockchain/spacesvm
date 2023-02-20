// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"github.com/MetalBlockchain/metalgo/database"
	"github.com/MetalBlockchain/metalgo/ids"
	"github.com/MetalBlockchain/metalgo/utils/set"
)

type Context struct {
	RecentBlockIDs  set.Set[ids.ID]
	RecentTxIDs     set.Set[ids.ID]
	RecentLoadUnits uint64

	Prices []uint64
	Costs  []uint64

	NextCost  uint64
	NextPrice uint64
}

type VM interface {
	Genesis() *Genesis
	IsBootstrapped() bool
	State() database.Database
	Mempool() Mempool
	GetStatelessBlock(ids.ID) (*StatelessBlock, error)
	ExecutionContext(currentTime int64, parent *StatelessBlock) (*Context, error)
	Verified(*StatelessBlock)
	Rejected(*StatelessBlock)
	Accepted(*StatelessBlock)
}
