package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/nullify"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
)

func createTestPaused(keeper *keeper.Keeper, ctx sdk.Context) types.Paused {
	item := types.Paused{}
	keeper.SetPaused(ctx, item)
	return item
}

func TestPausedGet(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
	item := createTestPaused(keeper, ctx)
	rst := keeper.GetPaused(ctx)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}
