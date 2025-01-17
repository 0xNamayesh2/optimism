package test

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// ClaimBuilder is a test utility to enable creating claims in a wide range of situations
type ClaimBuilder struct {
	require  *require.Assertions
	maxDepth int
	correct  types.TraceProvider
}

// NewClaimBuilder creates a new [ClaimBuilder].
func NewClaimBuilder(t *testing.T, maxDepth int, provider types.TraceProvider) *ClaimBuilder {
	return &ClaimBuilder{
		require:  require.New(t),
		maxDepth: maxDepth,
		correct:  provider,
	}
}

// CorrectTraceProvider returns a types.TraceProvider that provides the canonical trace.
func (c *ClaimBuilder) CorrectTraceProvider() types.TraceProvider {
	return c.correct
}

// CorrectClaimAtPosition returns the canonical claim at a specified position
func (c *ClaimBuilder) CorrectClaimAtPosition(pos types.Position) common.Hash {
	value, err := c.correct.Get(context.Background(), pos)
	c.require.NoError(err)
	return value
}

// CorrectPreState returns the pre-state (not hashed) required to execute the valid step at the specified trace index
func (c *ClaimBuilder) CorrectPreState(idx uint64) []byte {
	pos := types.NewPosition(c.maxDepth, int(idx))
	preimage, _, _, err := c.correct.GetStepData(context.Background(), pos)
	c.require.NoError(err)
	return preimage
}

// CorrectProofData returns the proof-data required to execute the valid step at the specified trace index
func (c *ClaimBuilder) CorrectProofData(idx uint64) []byte {
	pos := types.NewPosition(c.maxDepth, int(idx))
	_, proof, _, err := c.correct.GetStepData(context.Background(), pos)
	c.require.NoError(err)
	return proof
}

func (c *ClaimBuilder) CorrectOracleData(idx uint64) *types.PreimageOracleData {
	pos := types.NewPosition(c.maxDepth, int(idx))
	_, _, data, err := c.correct.GetStepData(context.Background(), pos)
	c.require.NoError(err)
	return data
}

func (c *ClaimBuilder) incorrectClaim(pos types.Position) common.Hash {
	return common.BigToHash(new(big.Int).SetUint64(pos.TraceIndex(c.maxDepth)))
}

func (c *ClaimBuilder) claim(pos types.Position, correct bool) common.Hash {
	if correct {
		return c.CorrectClaimAtPosition(pos)
	} else {
		return c.incorrectClaim(pos)
	}
}

func (c *ClaimBuilder) CreateRootClaim(correct bool) types.Claim {
	value := c.claim(types.NewPositionFromGIndex(1), correct)
	claim := types.Claim{
		ClaimData: types.ClaimData{
			Value:    value,
			Position: types.NewPosition(0, 0),
		},
	}
	return claim
}

func (c *ClaimBuilder) CreateLeafClaim(traceIndex uint64, correct bool) types.Claim {
	parentPos := types.NewPosition(c.maxDepth-1, 0)
	pos := types.NewPosition(c.maxDepth, int(traceIndex))
	return types.Claim{
		ClaimData: types.ClaimData{
			Value:    c.claim(pos, correct),
			Position: pos,
		},
		Parent: types.ClaimData{
			Value:    c.claim(parentPos, !correct),
			Position: parentPos,
		},
	}
}

func (c *ClaimBuilder) AttackClaim(claim types.Claim, correct bool) types.Claim {
	pos := claim.Position.Attack()
	return types.Claim{
		ClaimData: types.ClaimData{
			Value:    c.claim(pos, correct),
			Position: pos,
		},
		Parent:              claim.ClaimData,
		ParentContractIndex: claim.ContractIndex,
	}
}

func (c *ClaimBuilder) AttackClaimWithValue(claim types.Claim, value common.Hash) types.Claim {
	pos := claim.Position.Attack()
	return types.Claim{
		ClaimData: types.ClaimData{
			Value:    value,
			Position: pos,
		},
		Parent:              claim.ClaimData,
		ParentContractIndex: claim.ContractIndex,
	}
}

func (c *ClaimBuilder) DefendClaim(claim types.Claim, correct bool) types.Claim {
	pos := claim.Position.Defend()
	return types.Claim{
		ClaimData: types.ClaimData{
			Value:    c.claim(pos, correct),
			Position: pos,
		},
		Parent:              claim.ClaimData,
		ParentContractIndex: claim.ContractIndex,
	}
}

func (c *ClaimBuilder) DefendClaimWithValue(claim types.Claim, value common.Hash) types.Claim {
	pos := claim.Position.Defend()
	return types.Claim{
		ClaimData: types.ClaimData{
			Value:    value,
			Position: pos,
		},
		Parent:              claim.ClaimData,
		ParentContractIndex: claim.ContractIndex,
	}
}
