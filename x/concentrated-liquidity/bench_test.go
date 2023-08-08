package concentrated_liquidity_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v17/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity"
	clmath "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/math"
	clmodel "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v17/x/gamm/pool-models/balancer"
	gammmigration "github.com/osmosis-labs/osmosis/v17/x/gamm/types/migration"
)

type BenchTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *BenchTestSuite) createPosition(accountIndex int, poolId uint64, coin0, coin1 sdk.Coin, lowerTick, upperTick int64) {
	tokensDesired := sdk.NewCoins(coin0, coin1)

	_, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[accountIndex], tokensDesired, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
	if err != nil {
		// This can happen for ticks that map to the very small prices
		// e.g 2 * 10^(-18) ends up mapping to the same sqrt price
		fmt.Println("error creating position", err)
	}
}

func noError(b *testing.B, err error) {
	b.Helper()
	require.NoError(b, err)
}

func runBenchmark(b *testing.B, testFunc func(b *testing.B, s *BenchTestSuite, pool types.ConcentratedPoolExtension, largeSwapInCoin sdk.Coin, currentTick int64)) {
	b.Helper()
	// Notice we stop the timer to skip setup code.
	b.StopTimer()

	const (
		numberOfPositions              = 10000
		maxAmountDeposited             = int64(1_000_000_000_000)
		amountIn                       = "9999999999999999999"
		shouldCreateFullRangePositions = true
		shouldConcentrate              = true
		tickSpacing                    = 1
	)

	var (
		denom0               = DefaultCoin0.Denom
		denom1               = DefaultCoin1.Denom
		denomIn              = denom0
		numberOfPositionsInt = sdk.NewInt(numberOfPositions)
		maxAmountOfEachToken = sdk.NewInt(maxAmountDeposited).Mul(numberOfPositionsInt)
		seed                 = int64(1)
		defaultDenom0Asset   = balancer.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denom0, sdk.NewInt(1000000000)),
		}
		defaultDenom1Asset = balancer.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denom1, sdk.NewInt(1000000000)),
		}
		defaultPoolAssets = []balancer.PoolAsset{defaultDenom0Asset, defaultDenom1Asset}
	)

	rand.Seed(seed) //nolint:staticcheck // this can't have a security impact since it is part of a benchmark test

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := BenchTestSuite{}
		cleanup := s.SetupWithLevelDb()

		for _, acc := range s.TestAccs {
			err := simapp.FundAccount(s.App.BankKeeper, s.Ctx, acc, sdk.NewCoins(
				sdk.NewCoin(denom0, maxAmountOfEachToken),
				sdk.NewCoin(denom1, maxAmountOfEachToken),
				sdk.NewCoin("uosmo", maxAmountOfEachToken),
			))
			noError(b, err)
		}

		// Create a balancer pool
		gammPoolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, balancer.NewMsgCreateBalancerPool(s.TestAccs[0], balancer.PoolParams{
			SwapFee: sdk.MustNewDecFromStr("0.001"),
			ExitFee: sdk.ZeroDec(),
		}, defaultPoolAssets, ""))
		noError(b, err)

		// Create a cl pool.
		clPoolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(
			s.TestAccs[0], denom0, denom1, tickSpacing, sdk.MustNewDecFromStr("0.001"),
		))
		noError(b, err)

		clKeeper := s.App.ConcentratedLiquidityKeeper
		gammKeeper := s.App.GAMMKeeper

		// Create a link between the balancer and cl pool.
		record := gammmigration.BalancerToConcentratedPoolLink{BalancerPoolId: gammPoolId, ClPoolId: clPoolId}
		err = gammKeeper.ReplaceMigrationRecords(s.Ctx, []gammmigration.BalancerToConcentratedPoolLink{record})
		s.Require().NoError(err)

		_, err = gammKeeper.GetLinkedConcentratedPoolID(s.Ctx, gammPoolId)
		s.Require().NoError(err)

		// Create first position to set a price of 1 and tick of zero.
		tokenDesired0 := sdk.NewCoin(denom0, sdk.NewInt(100))
		tokenDesired1 := sdk.NewCoin(denom1, sdk.NewInt(100))
		tokensDesired := sdk.NewCoins(tokenDesired0, tokenDesired1)
		_, err = clKeeper.CreatePosition(s.Ctx, clPoolId, s.TestAccs[0], tokensDesired, sdk.ZeroInt(), sdk.ZeroInt(), types.MinInitializedTick, types.MaxTick)
		noError(b, err)

		pool, err := clKeeper.GetPoolById(s.Ctx, clPoolId)
		noError(b, err)

		// Zero by default, can configure by setting a specific position.
		currentTick := pool.GetCurrentTick()

		// Setup numberOfPositions positions at random ranges
		setupPositions := func() {
			for i := 0; i < numberOfPositions; i++ {
				var (
					lowerTick int64
					upperTick int64
				)

				if denomIn == denom0 {
					// Decreasing price so want to be below current tick
					// minTick <= lowerTick <= currentTick
					lowerTick = rand.Int63n(currentTick-types.MinInitializedTick+1) + types.MinInitializedTick
					// lowerTick <= upperTick <= currentTick
					upperTick = currentTick - rand.Int63n(int64(math.Abs(float64(currentTick-lowerTick))))
				} else {
					// Increasing price so want to be above current tick
					// currentTick <= lowerTick <= maxTick
					lowerTick = rand.Int63n(types.MaxTick-currentTick+1) + currentTick
					// lowerTick <= upperTick <= maxTick
					upperTick = types.MaxTick - rand.Int63n(int64(math.Abs(float64(types.MaxTick-lowerTick))))
				}

				// Normalize lowerTick to be a multiple of tickSpacing
				lowerTick = lowerTick + (tickSpacing - lowerTick%tickSpacing)
				// Normalize upperTick to be a multiple of tickSpacing
				upperTick = upperTick - upperTick%tickSpacing

				priceLowerTick, priceUpperTick, _, _, err := clmath.TicksToSqrtPrice(lowerTick, upperTick)
				noError(b, err)

				lowerTick, upperTick, err = cl.RoundTickToCanonicalPriceTick(
					lowerTick, upperTick, priceLowerTick, priceUpperTick, tickSpacing,
				)
				if err != nil {
					continue
				}

				tokenDesired0 := sdk.NewCoin(denom0, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
				tokenDesired1 := sdk.NewCoin(denom1, sdk.NewInt(rand.Int63n(maxAmountDeposited)))

				accountIndex := rand.Intn(len(s.TestAccs))
				s.createPosition(accountIndex, clPoolId, tokenDesired0, tokenDesired1, lowerTick, upperTick)
			}
		}

		createPosition := func(lowerTick, upperTick int64) {
			maxAmountDepositedFullRange := sdk.NewInt(maxAmountDeposited).MulRaw(5)
			tokenDesired0 := sdk.NewCoin(denom0, maxAmountDepositedFullRange)
			tokenDesired1 := sdk.NewCoin(denom1, maxAmountDepositedFullRange)
			tokensDesired := sdk.NewCoins(tokenDesired0, tokenDesired1)
			accountIndex := rand.Intn(len(s.TestAccs))
			account := s.TestAccs[accountIndex]
			err = simapp.FundAccount(s.App.BankKeeper, s.Ctx, account, tokensDesired)
			noError(b, err)
			s.createPosition(accountIndex, clPoolId, tokenDesired0, tokenDesired1, lowerTick, upperTick)
		}
		// Setup numberOfPositions full range positions for deeper liquidity.
		setupFullRangePositions := func() {
			for i := 0; i < numberOfPositions; i++ {
				lowerTick := types.MinInitializedTick
				upperTick := types.MaxTick
				createPosition(lowerTick, upperTick)
			}
		}

		// Setup numberOfPositions * 2 positions at random ranges around the current tick for deeper
		// liquidity.
		setupConcentratedPositions := func() {
			// Within 10 ticks of the current
			if tickSpacing <= 10 {
				for i := 0; i < numberOfPositions; i++ {
					lowerTick := currentTick - 10
					upperTick := currentTick + 10
					createPosition(lowerTick, upperTick)
				}
			}

			// Within 100 ticks of the current
			for i := 0; i < numberOfPositions; i++ {
				lowerTick := currentTick - 100
				upperTick := currentTick + 100
				lowerTick = lowerTick + (tickSpacing - lowerTick%tickSpacing)
				upperTick = upperTick - upperTick%tickSpacing
				createPosition(lowerTick, upperTick)
			}
		}

		setupPositions()
		if shouldCreateFullRangePositions {
			setupFullRangePositions()
		}
		if shouldConcentrate {
			setupConcentratedPositions()
		}

		swapAmountIn := sdk.MustNewDecFromStr(amountIn).TruncateInt()
		largeSwapInCoin := sdk.NewCoin(denomIn, swapAmountIn)
		// Commit so that the changes are propagated to IAVL.
		s.Commit()

		testFunc(b, &s, pool, largeSwapInCoin, currentTick)
		cleanup()
	}
}

func BenchmarkSwapExactAmountIn(b *testing.B) {
	b.Helper()
	runBenchmark(b, func(b *testing.B, s *BenchTestSuite, pool types.ConcentratedPoolExtension, largeSwapInCoin sdk.Coin, currentTick int64) {
		b.Helper()
		clKeeper := s.App.ConcentratedLiquidityKeeper

		liquidityNet, err := clKeeper.GetTickLiquidityNetInDirection(s.Ctx, pool.GetId(), largeSwapInCoin.Denom, sdk.NewInt(currentTick), sdk.Int{})
		noError(b, err)
		err = simapp.FundAccount(s.App.BankKeeper, s.Ctx, s.TestAccs[0], sdk.NewCoins(largeSwapInCoin))
		noError(b, err)

		b.StartTimer()

		// System under test
		_, err = clKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, largeSwapInCoin, DefaultCoin1.Denom, sdk.NewInt(1), pool.GetSpreadFactor(s.Ctx))
		b.StopTimer()
		noError(b, err)

		fmt.Println("current_tick", currentTick)
		fmt.Println("num_ticks_traversed", len(liquidityNet))
	})
}

func BenchmarkGetTickLiquidityNetInDirection(b *testing.B) {
	runBenchmark(b, func(b *testing.B, s *BenchTestSuite, pool types.ConcentratedPoolExtension, largeSwapInCoin sdk.Coin, currentTick int64) {
		b.Helper()
		clKeeper := s.App.ConcentratedLiquidityKeeper

		b.StartTimer()

		// System under test
		liquidityNet, err := clKeeper.GetTickLiquidityNetInDirection(s.Ctx, pool.GetId(), largeSwapInCoin.Denom, sdk.NewInt(currentTick), sdk.Int{})
		b.StopTimer()
		noError(b, err)

		fmt.Println("current_tick", currentTick)
		fmt.Println("num_ticks_traversed", len(liquidityNet))
	})
}

func BenchmarkGetTickLiquidityForFullRange(b *testing.B) {
	runBenchmark(b, func(b *testing.B, s *BenchTestSuite, pool types.ConcentratedPoolExtension, largeSwapInCoin sdk.Coin, currentTick int64) {
		b.Helper()
		clKeeper := s.App.ConcentratedLiquidityKeeper

		b.StartTimer()

		// System under test
		liquidityNet, err := clKeeper.GetTickLiquidityForFullRange(s.Ctx, pool.GetId())
		b.StopTimer()
		noError(b, err)

		fmt.Println("current_tick", currentTick)
		fmt.Println("num_ticks_traversed", len(liquidityNet))
	})
}
