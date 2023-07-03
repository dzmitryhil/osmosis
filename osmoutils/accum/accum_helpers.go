package accum

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

// initOrUpdatePosition creates a new position or override an existing position
// at accumulator's current value with a specific number of shares and unclaimed rewards
func initOrUpdatePosition(accum AccumulatorObject, accumulatorValuePerShare sdk.DecCoins, index string, numShareUnits sdk.Dec, unclaimedRewardsTotal sdk.DecCoins, options *Options) {
	position := Record{
		NumShares:             numShareUnits,
		AccumValuePerShare:    accumulatorValuePerShare,
		UnclaimedRewardsTotal: unclaimedRewardsTotal,
		Options:               options,
	}
	osmoutils.MustSet(accum.store, FormatPositionPrefixKey(accum.name, index), &position)
}

// Gets addr's current position from store
func GetPosition(accum AccumulatorObject, name string) (Record, error) {
	position := Record{}
	found, err := osmoutils.Get(accum.store, FormatPositionPrefixKey(accum.name, name), &position)
	if err != nil {
		return Record{}, err
	}
	if !found {
		return Record{}, NoPositionError{name}
	}

	return position, nil
}

// Gets total unclaimed rewards, including existing and newly accrued unclaimed rewards
func GetTotalRewards(accum AccumulatorObject, position Record) sdk.DecCoins {
	totalRewards := position.UnclaimedRewardsTotal

	// TODO: add a check that accum.value is greater than position.InitAccumValue
	// fmt.Println("accum.valuePerShare", accum.valuePerShare)
	// fmt.Println("position.AccumValuePerShare", position.AccumValuePerShare)
	globalVsPositionDiff, anyNegative := accum.valuePerShare.SafeSub(position.AccumValuePerShare)
	if anyNegative {
		panic(fmt.Sprintf("accum.valuePerShare must be greater than position.AccumValuePerShare.\n global: %v\n position: %v",
			accum.valuePerShare, position.AccumValuePerShare))
	}
	accumulatorRewards := globalVsPositionDiff.MulDec(position.NumShares)
	totalRewards = totalRewards.Add(accumulatorRewards...)

	return totalRewards
}
