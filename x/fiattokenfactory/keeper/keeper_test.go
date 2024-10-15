// Copyright 2024 Circle Internet Group, Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	"github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestKeeperLogger(t *testing.T) {
	t.Parallel()
	k, ctx := keeper.FiatTokenfactoryKeeper()
	l := k.Logger(ctx)
	require.NotNil(t, l)
}

func TestSendRestrictionsFn_NotUsingUSDC(t *testing.T) {
	k, ctx := keeper.FiatTokenfactoryKeeper()
	k.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
	fromAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	toAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	amounts := sdk.Coins{sdk.NewInt64Coin("blahblah", 10)}

	newToAddress, err := k.SendRestrictionFn(ctx, fromAddress, toAddress, amounts)

	require.Nil(t, err)
	require.Equal(t, toAddress, newToAddress)
}

func TestSendRestrictionsFn_Paused(t *testing.T) {
	k, ctx := keeper.FiatTokenfactoryKeeper()
	k.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
	k.SetPaused(ctx, types.Paused{Paused: true})
	fromAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	toAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	amounts := sdk.Coins{sdk.NewInt64Coin("uusdc", 10)}

	newToAddress, err := k.SendRestrictionFn(ctx, fromAddress, toAddress, amounts)

	require.ErrorIs(t, err, types.ErrPaused)
	require.Equal(t, toAddress, newToAddress)
}

func TestSendRestrictionsFn_FromBlacklisted(t *testing.T) {
	k, ctx := keeper.FiatTokenfactoryKeeper()
	k.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
	k.SetPaused(ctx, types.Paused{Paused: false})

	fromAccount := sample.TestAccount()
	fromAddress := sdk.MustAccAddressFromBech32(fromAccount.Address)
	k.SetBlacklisted(ctx, types.Blacklisted{AddressBz: fromAccount.AddressBz})
	toAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	amounts := sdk.Coins{sdk.NewInt64Coin("uusdc", 10)}

	newToAddress, err := k.SendRestrictionFn(ctx, fromAddress, toAddress, amounts)

	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.Equal(t, toAddress, newToAddress)
}

func TestSendRestrictionsFn_ToBlacklisted(t *testing.T) {
	k, ctx := keeper.FiatTokenfactoryKeeper()
	k.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
	k.SetPaused(ctx, types.Paused{Paused: false})

	fromAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	toAccount := sample.TestAccount()
	toAddress := sdk.MustAccAddressFromBech32(toAccount.Address)
	k.SetBlacklisted(ctx, types.Blacklisted{AddressBz: toAccount.AddressBz})
	amounts := sdk.Coins{sdk.NewInt64Coin("uusdc", 10)}

	newToAddress, err := k.SendRestrictionFn(ctx, fromAddress, toAddress, amounts)

	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.Equal(t, toAddress, newToAddress)
}

func TestSendRestrictionsFn_GranteeBlacklisted(t *testing.T) {
	k, ctx := keeper.FiatTokenfactoryKeeper()
	k.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
	k.SetPaused(ctx, types.Paused{Paused: false})

	fromAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	toAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	amounts := sdk.Coins{sdk.NewInt64Coin("uusdc", 10)}

	granteeAccount := sample.TestAccount()
	k.SetBlacklisted(ctx, types.Blacklisted{AddressBz: granteeAccount.AddressBz})

	var grantees []string
	grantees = append(grantees, granteeAccount.Address)

	newToAddress, err := k.SendRestrictionFn(ctx.WithValue(types.GranteeKey, grantees), fromAddress, toAddress, amounts)

	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.Equal(t, toAddress, newToAddress)
}

func TestSendRestrictionsFn_Bech32mGranteeBlacklisted(t *testing.T) {
	k, ctx := keeper.FiatTokenfactoryKeeper()
	k.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
	k.SetPaused(ctx, types.Paused{Paused: false})

	fromAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	toAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	amounts := sdk.Coins{sdk.NewInt64Coin("uusdc", 10)}

	granteeAccount := sample.TestAccountBech32m()
	k.SetBlacklisted(ctx, types.Blacklisted{AddressBz: granteeAccount.AddressBz})

	var grantees []string
	grantees = append(grantees, granteeAccount.Address)

	newToAddress, err := k.SendRestrictionFn(ctx.WithValue(types.GranteeKey, grantees), fromAddress, toAddress, amounts)

	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.Equal(t, toAddress, newToAddress)
}

func TestSendRestrictionsFn_USDCNotRestricted(t *testing.T) {
	k, ctx := keeper.FiatTokenfactoryKeeper()
	k.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
	k.SetPaused(ctx, types.Paused{Paused: false})
	fromAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	toAddress := sdk.MustAccAddressFromBech32(sample.TestAccount().Address)
	amounts := sdk.Coins{sdk.NewInt64Coin("uusdc", 10)}

	newToAddress, err := k.SendRestrictionFn(ctx, fromAddress, toAddress, amounts)

	require.NoError(t, err)
	require.Equal(t, toAddress, newToAddress)
}

func TestValidatePrivileges_InvalidAddress(t *testing.T) {
	k, ctx := keeper.FiatTokenfactoryKeeper()

	err := k.ValidatePrivileges(ctx, "malformed bech32 address")
	require.Error(t, err)
}

func TestValidatePrivileges_NonPrivilegedAddress(t *testing.T) {
	owner := sample.TestAccount()
	blacklister := sample.TestAccount()
	masterMinter := sample.TestAccount()
	pauser := sample.TestAccount()
	k, ctx := keeper.FiatTokenfactoryKeeper()
	k.SetOwner(ctx, types.Owner{Address: owner.Address})
	k.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})
	k.SetMasterMinter(ctx, types.MasterMinter{Address: masterMinter.Address})
	k.SetPauser(ctx, types.Pauser{Address: pauser.Address})

	newAddress := sample.TestAccount()
	err := k.ValidatePrivileges(ctx, newAddress.Address)

	require.NoError(t, err)
}

func TestValidatePrivileges_PrivilegedAddress(t *testing.T) {
	owner := sample.TestAccount()
	blacklister := sample.TestAccount()
	masterMinter := sample.TestAccount()
	pauser := sample.TestAccount()
	k, ctx := keeper.FiatTokenfactoryKeeper()
	k.SetOwner(ctx, types.Owner{Address: owner.Address})
	k.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})
	k.SetMasterMinter(ctx, types.MasterMinter{Address: masterMinter.Address})
	k.SetPauser(ctx, types.Pauser{Address: pauser.Address})
	address := []sample.Account{owner, blacklister, masterMinter, pauser}

	for _, ad := range address {
		err := k.ValidatePrivileges(ctx, ad.Address)
		require.ErrorIs(t, err, types.ErrAlreadyPrivileged)
	}
}
