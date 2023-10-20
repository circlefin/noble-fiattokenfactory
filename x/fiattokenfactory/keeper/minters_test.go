package keeper_test

import (
	"strconv"
	"testing"

	keepertest "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/nullify"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNMinters(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Minters {
	items := make([]types.Minters, n)
	for i := range items {
		items[i].Address = strconv.Itoa(i)

		keeper.SetMinters(ctx, items[i])
	}
	return items
}

func TestMintersGet(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
	items := createNMinters(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetMinters(ctx,
			item.Address,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestMintersRemove(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
	items := createNMinters(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveMinters(ctx,
			item.Address,
		)
		_, found := keeper.GetMinters(ctx,
			item.Address,
		)
		require.False(t, found)
	}
}

func TestMintersGetAll(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
	items := createNMinters(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllMinters(ctx)),
	)
}
