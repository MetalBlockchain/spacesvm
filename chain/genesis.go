// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/inconshreveable/log15"
)

type Airdrop struct {
	// Address strings are hex-formatted common.Address
	Address common.Address `serialize:"true" json:"address"`
}

type CustomAllocation struct {
	// Address strings are hex-formatted common.Address
	Address common.Address `serialize:"true" json:"address"`
	Balance uint64         `serialize:"true" json:"balance"`
}

type Genesis struct {
	Magic uint64 `serialize:"true" json:"magic"`

	// Tx params
	BaseTxUnits uint64 `serialize:"true" json:"baseTxUnits"`

	// SetTx params
	ValueUnitSize uint64 `serialize:"true" json:"valueUnitSize"`
	MaxValueSize  uint64 `serialize:"true" json:"maxValueSize"`

	// Claim Params
	ClaimFeeMultiplier   uint64 `serialize:"true" json:"claimFeeMultiplier"`
	ClaimTier3Multiplier uint64 `serialize:"true" json:"claimTier3Multiplier"`
	ClaimTier2Size       uint64 `serialize:"true" json:"claimTier2Size"`
	ClaimTier2Multiplier uint64 `serialize:"true" json:"claimTier2Multiplier"`
	ClaimTier1Size       uint64 `serialize:"true" json:"claimTier1Size"`
	ClaimTier1Multiplier uint64 `serialize:"true" json:"claimTier1Multiplier"`

	// Lifeline Params
	PrefixRenewalDiscount uint64 `serialize:"true" json:"prefixRenewalDiscount"`

	// Reward Params
	ClaimReward        uint64 `serialize:"true" json:"claimReward"`
	LifelineUnitReward uint64 `serialize:"true" json:"lifelineUnitReward"`

	// Mining Reward (% of min required fee)
	LotteryRewardMultipler uint64 `serialize:"true" json:"lotteryRewardMultipler"`
	LotteryRewardDivisor   uint64 `serialize:"true" json:"lotteryRewardDivisor"`

	// Fee Mechanism Params
	LookbackWindow int64  `serialize:"true" json:"lookbackWindow"`
	BlockTarget    int64  `serialize:"true" json:"blockTarget"`
	TargetUnits    uint64 `serialize:"true" json:"targetUnits"`
	MinPrice       uint64 `serialize:"true" json:"minPrice"`
	MinBlockCost   uint64 `serialize:"true" json:"minBlockCost"`

	// Allocations
	CustomAllocation []*CustomAllocation `serialize:"true" json:"customAllocation"`
	AirdropHash      string              `serialize:"true" json:"airdropHash"`
	AirdropUnits     uint64              `serialize:"true" json:"airdropUnits"`
}

func DefaultGenesis() *Genesis {
	return &Genesis{
		// Tx params
		BaseTxUnits: 10,

		// SetTx params
		ValueUnitSize: 256,             // 256B
		MaxValueSize:  128 * units.KiB, // (500 Units)

		// Claim Params
		ClaimFeeMultiplier:   5,
		ClaimTier3Multiplier: 1,
		ClaimTier2Size:       36,
		ClaimTier2Multiplier: 5,
		ClaimTier1Size:       12,
		ClaimTier1Multiplier: 25,

		// Lifeline Params
		PrefixRenewalDiscount: 5,

		// Reward Params
		ClaimReward:        60 * 60 * 24 * 15, // 15 Days
		LifelineUnitReward: 60 * 60 * 6,       // 6 Hours Per Fee Unit (1 ms of work)

		// Lottery Reward (80% of tx.FeeUnits() * block.Price)
		LotteryRewardMultipler: 8,
		LotteryRewardDivisor:   10,

		// Fee Mechanism Params
		LookbackWindow: 60,            // 60 Seconds
		BlockTarget:    1,             // 1 Block per Second
		TargetUnits:    10 * 512 * 60, // 5012 Units Per Block (~1.2MB of SetTx)
		MinPrice:       1,             // (50 for easiest claim)
		MinBlockCost:   0,             // Minimum Unit Overhead
	}
}

func (g *Genesis) StatefulBlock() *StatefulBlock {
	return &StatefulBlock{
		Price: g.MinPrice,
		Cost:  g.MinBlockCost,
	}
}

func (g *Genesis) Verify() error {
	if g.Magic == 0 {
		return ErrInvalidMagic
	}
	return nil
}

func (g *Genesis) Load(db database.KeyValueWriter, airdropData []byte) error {
	if len(g.AirdropHash) > 0 {
		h := common.BytesToHash(crypto.Keccak256(airdropData)).Hex()
		if g.AirdropHash != h {
			return fmt.Errorf("expected standard allocation %s but got %s", g.AirdropHash, h)
		}

		standardAllocation := []*Airdrop{}
		if err := json.Unmarshal(airdropData, &standardAllocation); err != nil {
			return err
		}

		for _, alloc := range standardAllocation {
			if err := SetBalance(db, alloc.Address, g.AirdropUnits); err != nil {
				return fmt.Errorf("%w: addr=%s, bal=%d", err, alloc.Address, g.AirdropUnits)
			}
		}
		log.Debug(
			"applied airdrop allocation",
			"hash", h, "addrs", len(standardAllocation), "balance", g.AirdropUnits,
		)
	}

	// Do custom allocation last in case an address shows up in standard
	// allocation
	for _, alloc := range g.CustomAllocation {
		if err := SetBalance(db, alloc.Address, alloc.Balance); err != nil {
			return fmt.Errorf("%w: addr=%s, bal=%d", err, alloc.Address, alloc.Balance)
		}
		log.Debug("applied custom allocation", "addr", alloc.Address, "balance", alloc.Balance)
	}
	return nil
}
