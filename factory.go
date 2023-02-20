// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"github.com/MetalBlockchain/metalgo/snow"
	"github.com/MetalBlockchain/metalgo/vms"
	"github.com/MetalBlockchain/spacesvm/vm"
)

var _ vms.Factory = &Factory{}

type Factory struct{}

func (f *Factory) New(*snow.Context) (interface{}, error) { return &vm.VM{}, nil }
