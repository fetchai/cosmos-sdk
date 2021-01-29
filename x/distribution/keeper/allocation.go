package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"

	//tmtypes "github.com/tendermint/tendermint/types"
)

// AllocateTokens handles distribution of the collected fees
func (k Keeper) AllocateTokens(
	ctx sdk.Context, sumPreviousPrecommitPower, totalPreviousPower int64,
	previousProposer sdk.ConsAddress, req abci.RequestBeginBlock,
) {

	logger := k.Logger(ctx)
	previousVotes := req.LastCommitInfo.GetVotes()
	//entropy := req.Header.GetEntropy()

	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer). feesCollected includes the
	// block reward and transaction fees.
	feeCollector := k.supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := feeCollector.GetCoins()
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	// DEBUG: Total block rewards
	logger.Error("### [AllocateTokens]", "collector", k.feeCollectorName)
	for _, feeDemon := range feesCollected {
		logger.Error("### [AllocateTokens]", "feeDemon", feeDemon)
	}

	// transfer collected fees to the distribution module account
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, feesCollectedInt)
	if err != nil {
		panic(err)
	}

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/cosmos/cosmos-sdk/issues/2906#issuecomment-441867634
	feePool := k.GetFeePool(ctx)
	if totalPreviousPower == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		k.SetFeePool(ctx, feePool)
		return
	}

	remaining := feesCollected

	// If block contains entropy then use proportion of block reward to reward to validators which successfully
	// completed the dkg. If block does not contain entropy then this proportion of the fees collected goes into
	// the community pool
	//beaconRewardMultiplier := k.GetBeaconReward(ctx)
	//vals := []*tmtypes.Validator{}
	//if len(entropy.GroupSignature) != 0 {
	//	// Get all validators (including those than have been jailed)
	//	k.stakingKeeper.IterateLastValidators(ctx, func(index int64, val exported.ValidatorI) bool {
	//		vals = append(vals, tmtypes.NewValidator(val.GetConsPubKey(), val.GetConsensusPower()))
	//		return false
	//	})
	//
	//	// Pay beacon rewards to those in qual
	//	valSet := &tmtypes.ValidatorSet{}
	//	err = valSet.UpdateWithChangeSet(vals)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	for _, index := range entropy.SuccessfulVals {
	//		addr, val := valSet.GetByIndex(int(index))
	//		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, sdk.ConsAddress(addr))
	//		powerFraction := sdk.NewDec(val.VotingPower).QuoTruncate(sdk.NewDec(valSet.TotalVotingPower()))
	//		reward := feesCollected.MulDecTruncate(beaconRewardMultiplier).MulDecTruncate(powerFraction)
	//		k.AllocateTokensToValidator(ctx, validator, reward)
	//		logger.Debug("Beacon reward allocated", "val", fmt.Sprintf("%X", addr), "amount", reward.String())
	//		remaining = remaining.Sub(reward)
	//	}
	//}

	logger.Error("### [AllocateTokens]", "sumPreviousPrecommitPower", sumPreviousPrecommitPower)
	logger.Error("### [AllocateTokens]", "totalPreviousPower", totalPreviousPower)

	// calculate fraction votes
	previousFractionVotes := sdk.NewDec(sumPreviousPrecommitPower).Quo(sdk.NewDec(totalPreviousPower))

	// calculate previous proposer reward
	baseProposerReward := k.GetBaseProposerReward(ctx)
	bonusProposerReward := k.GetBonusProposerReward(ctx)
	proposerMultiplier := baseProposerReward.Add(bonusProposerReward.MulTruncate(previousFractionVotes))
	proposerReward := feesCollected.MulDecTruncate(proposerMultiplier)

	logger.Error("### [AllocateTokens]", "feesCollected", feesCollected)
	logger.Error("### [AllocateTokens]", "previousProposer", previousProposer)
	logger.Error("### [AllocateTokens]", "bonusProposerReward", bonusProposerReward)
	logger.Error("### [AllocateTokens]", "previousFractionVotes", previousFractionVotes)
	logger.Error("### [AllocateTokens]", "proposerMultiplier", proposerMultiplier)
	logger.Error("### [AllocateTokens]", "proposerReward", proposerReward)

	// pay previous proposer
	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, previousProposer)

	if proposerValidator != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposerReward,
				sdk.NewAttribute(sdk.AttributeKeyAmount, proposerReward.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, proposerValidator.GetOperator().String()),
			),
		)

		k.AllocateTokensToValidator(ctx, proposerValidator, proposerReward)
		remaining = remaining.Sub(proposerReward)
	} else {
		// previous proposer can be unknown if say, the unbonding period is 1 block, so
		// e.g. a validator undelegates at block X, it's removed entirely by
		// block X+1's endblock, then X+2 we need to refer to the previous
		// proposer for X+1, but we've forgotten about them.
		logger.Error(fmt.Sprintf(
			"WARNING: Attempt to allocate proposer rewards to unknown proposer %s. "+
				"This should happen only if the proposer unbonded completely within a single block, "+
				"which generally should not happen except in exceptional circumstances (or fuzz testing). "+
				"We recommend you investigate immediately.",
			previousProposer.String()))
	}

	// calculate fraction allocated to validators
	communityTax := k.GetCommunityTax(ctx)
	voteMultiplier := sdk.OneDec().Sub(proposerMultiplier).Sub(communityTax) //.Sub(beaconRewardMultiplier)

	logger.Error("### [AllocateTokens]", "oneDec", sdk.OneDec())
	logger.Error("### [AllocateTokens]", "communityTax", communityTax)
	logger.Error("### [AllocateTokens]", "voteMultiplier", voteMultiplier)

	// allocate tokens proportionally to voting power
	// TODO consider parallelizing later, ref https://github.com/cosmos/cosmos-sdk/pull/3099#discussion_r246276376
	for idx, vote := range previousVotes {
		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)

		// TODO consider microslashing for missing votes.
		// ref https://github.com/cosmos/cosmos-sdk/issues/2525#issuecomment-430838701
		powerFraction := sdk.NewDec(vote.Validator.Power).QuoTruncate(sdk.NewDec(totalPreviousPower))
		reward := feesCollected.MulDecTruncate(voteMultiplier).MulDecTruncate(powerFraction)

		logger.Error("### [AllocateTokens]", "voteIdx", idx, "powerFraction", powerFraction, "reward", reward)
		k.AllocateTokensToValidator(ctx, validator, reward)

		remaining = remaining.Sub(reward)
	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining...)
	logger.Error("### [AllocateTokens]", "communityPool", remaining, "total", feePool.CommunityPool)

	k.SetFeePool(ctx, feePool)
}

// AllocateTokensToValidator allocate tokens to a particular validator, splitting according to commission
func (k Keeper) AllocateTokensToValidator(ctx sdk.Context, val exported.ValidatorI, tokens sdk.DecCoins) {
	logger := k.Logger(ctx)

	// split tokens between validator and delegators according to commission
	commission := tokens.MulDec(val.GetCommission())
	shared := tokens.Sub(commission)

	// update current commission
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	currentCommission := k.GetValidatorAccumulatedCommission(ctx, val.GetOperator())
	currentCommission = currentCommission.Add(commission...)

	logger.Error("### [AllocateTokensToValidator]", "commission", commission, "total", currentCommission)

	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), currentCommission)

	// update current rewards
	currentRewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())
	currentRewards.Rewards = currentRewards.Rewards.Add(shared...)

	logger.Error("### [AllocateTokensToValidator]", "shared", shared, "total", currentRewards.Rewards)

	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), currentRewards)

	// update outstanding rewards
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewards,
			sdk.NewAttribute(sdk.AttributeKeyAmount, tokens.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
	outstanding = outstanding.Add(tokens...)

	logger.Error("### [AllocateTokensToValidator]", "outstanding", tokens, "total", outstanding)

	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)
}
