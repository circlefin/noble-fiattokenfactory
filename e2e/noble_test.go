// Copyright 2024 Circle Internet Group, Inc. All rights reserved.
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

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"

	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer/hermes"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// run `make local-image`to rebuild updated binary before running test
func TestNobleChain(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	var gw genesisWrapper

	numValidators, numFullNodes := 1, 0
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", 2, 1, false, true),
		{
			Name:          "gaia",
			Version:       "v14.1.0",
			NumValidators: &numValidators,
			NumFullNodes:  &numFullNodes,
			ChainConfig: ibc.ChainConfig{
				ChainID: "cosmoshub-4",
			},
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	gaia := chains[1].(*cosmos.CosmosChain)

	gw.chain = chains[0].(*cosmos.CosmosChain)
	noble := gw.chain

	rly := interchaintest.NewBuiltinRelayerFactory(
		ibc.Hermes,
		zaptest.NewLogger(t),
	).Build(t, client, network).(*hermes.Relayer)

	ic := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(rly, "hermes").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  gaia,
			Relayer: rly,
			Path:    "transfer",
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	gaiaWallets := interchaintest.GetAndFundTestUsers(t, ctx, "receiver", math.NewInt(1_000_000), gaia, gaia)

	t.Run("fiat-tokenfactory", func(t *testing.T) {
		t.Parallel()
		nobleTokenfactory_e2e(t, ctx, "fiat-tokenfactory", denomMetadataDrachma.Base, noble, gaia, gw.fiatTfRoles, gw.extraWallets, gaiaWallets)
	})

	err = rly.StartRelayer(ctx, eRep, "transfer")
	require.NoError(t, err, "failed to start relayer")
}

func nobleTokenfactory_e2e(t *testing.T, ctx context.Context, tokenfactoryModName, mintingDenom string, noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, roles NobleRoles, extraWallets ExtraWallets, gaiaWallets []ibc.Wallet) {
	nobleValidator := noble.Validators[0]

	_, err := nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-master-minter", roles.MasterMinter.FormattedAddress(),
	)
	require.Error(t, err, "succeeded to execute update master minter tx by invalid owner")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-owner", roles.Owner2.FormattedAddress(),
	)
	require.Error(t, err, "succeeded to execute update owner tx by invalid owner")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner.KeyName(),
		tokenfactoryModName, "update-owner", roles.Owner2.FormattedAddress(),
	)
	require.NoError(t, err, "failed to execute update owner tx")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-master-minter", roles.MasterMinter.FormattedAddress(),
	)
	require.Error(t, err, "succeeded to execute update master minter tx by pending owner")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "accept-owner",
	)
	require.NoError(t, err, "failed to execute tx to accept ownership")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner.KeyName(),
		tokenfactoryModName, "update-master-minter", roles.MasterMinter.FormattedAddress(),
	)
	require.Error(t, err, "succeeded to execute update master minter tx by prior owner")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-master-minter", roles.MasterMinter.FormattedAddress(),
	)
	require.NoError(t, err, "failed to execute update master minter tx")

	_, err = nobleValidator.ExecTx(ctx, roles.MasterMinter.KeyName(),
		tokenfactoryModName, "configure-minter-controller", roles.MinterController.FormattedAddress(), roles.Minter.FormattedAddress(),
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, roles.MinterController.KeyName(),
		tokenfactoryModName, "configure-minter", roles.Minter.FormattedAddress(), "1000"+mintingDenom,
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User.FormattedAddress(), "200"+mintingDenom,
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err := noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equalf(t, math.NewInt(200), userBalance, "failed to mint %s to user", mintingDenom)

	// Fund gaia wallets with tokens to prepare for IBC tests
	testIBCTransferSucceed(t, ctx, mintingDenom, noble, gaia, extraWallets.User, gaiaWallets[0])
	testIBCTransferSucceed(t, ctx, mintingDenom, noble, gaia, extraWallets.User, gaiaWallets[1])

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-blacklister", roles.Blacklister.FormattedAddress(),
	)
	require.NoError(t, err, "failed to set blacklister")

	_, err = nobleValidator.ExecTx(ctx, roles.Blacklister.KeyName(),
		tokenfactoryModName, "blacklist", extraWallets.User.FormattedAddress(),
	)
	require.NoError(t, err, "failed to blacklist user address")

	gaiaWalletBech32Addr, err := sdk.Bech32ifyAddressBytes(noble.Config().Bech32Prefix, gaiaWallets[0].Address())
	require.NoError(t, err, "failed to convert gaia wallet address")

	_, err = nobleValidator.ExecTx(ctx, roles.Blacklister.KeyName(),
		tokenfactoryModName, "blacklist", gaiaWalletBech32Addr,
	)
	require.NoError(t, err, "failed to blacklist gaia wallet address")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User.FormattedAddress(), "100"+mintingDenom,
	)
	require.Error(t, err, "successfully executed mint to blacklisted user tx")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, math.NewInt(100), userBalance, "user balance should not have incremented while blacklisted")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User2.FormattedAddress(), "100"+mintingDenom,
	)
	require.NoError(t, err, "failed to execute mint to user2 tx")

	err = nobleValidator.BankSend(ctx, extraWallets.User2.KeyName(), ibc.WalletAmount{
		Address: extraWallets.User.FormattedAddress(),
		Denom:   mintingDenom,
		Amount:  math.NewInt(50),
	})
	require.Error(t, err, "The tx to a blacklisted user should not have been successful")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, math.NewInt(100), userBalance, "user balance should not have incremented while blacklisted")

	// paying gas in minting denom fails when spender is blacklisted
	nobleStakeWallets := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.NewInt(1_000), noble, noble)
	_, err = nobleValidator.ExecTx(ctx, nobleStakeWallets[0].KeyName(),
		"bank", "send", nobleStakeWallets[0].KeyName(), extraWallets.User.FormattedAddress(), "100token",
	)
	_, err = nobleValidator.ExecTx(ctx, extraWallets.User.KeyName(),
		"bank", "send", extraWallets.User.KeyName(), extraWallets.Alice.FormattedAddress(), "10token", "--fees", "1"+mintingDenom,
	)
	require.ErrorContains(t, err, "blacklisted", "failed to block spending gas in minting denom")

	// IBC transfer from blacklisted account, from noble to gaia
	testIBCTransferFail(t, ctx, mintingDenom, noble, gaia, extraWallets.User, extraWallets.User2, "blacklisted")
	// IBC transfer to blacklisted account, from noble to gaia
	testIBCTransferFail(t, ctx, mintingDenom, noble, gaia, extraWallets.User2, extraWallets.User, "blacklisted")

	// IBC transfer from blacklisted account, from gaia to noble
	testReverseIBCTransferFail(t, ctx, mintingDenom, gaia, noble, gaiaWallets[0], extraWallets.User2, "not found")

	// IBC transfer to blacklisted account, from gaia to noble
	testReverseIBCTransferFail(t, ctx, mintingDenom, gaia, noble, gaiaWallets[1], extraWallets.User, "not found")

	// authz send to blacklisted account
	testAuthZSendFail(t, ctx, nobleValidator, mintingDenom, noble, extraWallets.User2, extraWallets.User, extraWallets.Alice, "blacklisted", false)
	// authz send from blacklisted account
	testAuthZSendFail(t, ctx, nobleValidator, mintingDenom, noble, extraWallets.User, extraWallets.User2, extraWallets.Alice, "blacklisted", false)
	// authz send with blacklisted grantee
	testAuthZSendFail(t, ctx, nobleValidator, mintingDenom, noble, extraWallets.User2, extraWallets.Alice, extraWallets.User, "blacklisted", false)
	// authz ibc transfer to blacklisted account
	testAuthZIBCTransferFail(t, ctx, nobleValidator, mintingDenom, noble, gaia, extraWallets.User2, extraWallets.User, extraWallets.Alice, "blacklisted", false)
	// authz ibc transfer from blacklisted account
	testAuthZIBCTransferFail(t, ctx, nobleValidator, mintingDenom, noble, gaia, extraWallets.User, extraWallets.User2, extraWallets.Alice, "blacklisted", false)
	// authz ibc transfer with blacklisted grantee
	testAuthZIBCTransferFail(t, ctx, nobleValidator, mintingDenom, noble, gaia, extraWallets.User2, extraWallets.Alice, extraWallets.User, "blacklisted", false)

	err = nobleValidator.BankSend(ctx, extraWallets.User2.KeyName(), ibc.WalletAmount{
		Address: extraWallets.User.FormattedAddress(),
		Denom:   "token",
		Amount:  math.NewInt(100),
	})
	require.NoError(t, err, "The tx should have been successfull as that is not the minting denom")

	_, err = nobleValidator.ExecTx(ctx, roles.Blacklister.KeyName(),
		tokenfactoryModName, "unblacklist", extraWallets.User.FormattedAddress(),
	)
	require.NoError(t, err, "failed to unblacklist user address")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User.FormattedAddress(), "100"+mintingDenom,
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, math.NewInt(200), userBalance, "user balance should have increased now that they are no longer blacklisted")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", roles.Minter.FormattedAddress(), "100"+mintingDenom,
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	minterBalance, err := noble.GetBalance(ctx, roles.Minter.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get minter balance")
	require.Equal(t, math.NewInt(100), minterBalance, "minter balance should have increased")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "burn", "10"+mintingDenom,
	)
	require.NoError(t, err, "failed to execute burn tx")

	minterBalance, err = noble.GetBalance(ctx, roles.Minter.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get minter balance")
	require.Equal(t, math.NewInt(90), minterBalance, "minter balance should have decreased because tokens were burned")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-pauser", roles.Pauser.FormattedAddress(),
	)
	require.NoError(t, err, "failed to update pauser")

	_, err = nobleValidator.ExecTx(ctx, roles.MasterMinter.KeyName(),
		tokenfactoryModName, "configure-minter-controller", roles.MinterController2.FormattedAddress(), extraWallets.User.FormattedAddress())
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, roles.MinterController2.KeyName(),
		tokenfactoryModName, "configure-minter", extraWallets.User.FormattedAddress(), "1000"+mintingDenom)
	require.NoError(t, err, "failed to execute configure minter tx")

	// -- chain paused --

	_, err = nobleValidator.ExecTx(ctx, roles.Pauser.KeyName(),
		tokenfactoryModName, "pause",
	)
	require.NoError(t, err, "failed to pause mints")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User.FormattedAddress(), "100"+mintingDenom,
	)
	require.Error(t, err, "successfully executed mint to user tx while chain is paused")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, math.NewInt(200), userBalance, "user balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, extraWallets.User.KeyName(),
		"bank", "send", extraWallets.User.FormattedAddress(), extraWallets.Alice.FormattedAddress(), "100"+mintingDenom,
	)
	require.Error(t, err, "transaction was successful while chain is paused")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, math.NewInt(200), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err := noble.GetBalance(ctx, extraWallets.Alice.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get alice balance")

	require.Equal(t, math.ZeroInt(), aliceBalance, "alice balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "burn", "10"+mintingDenom,
	)
	require.Error(t, err, "successfully executed burn tx while chain is paused")
	require.Equal(t, math.NewInt(90), minterBalance, "this burn should not have been successful because the chain is paused")

	_, err = nobleValidator.ExecTx(ctx, roles.MasterMinter.KeyName(),
		tokenfactoryModName, "configure-minter-controller", roles.MinterController2.FormattedAddress(), extraWallets.User.FormattedAddress())

	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, roles.MinterController2.KeyName(),
		tokenfactoryModName, "configure-minter", extraWallets.User.FormattedAddress(), "1000"+mintingDenom)
	require.Error(t, err, "failed to block configure minter tx when chain is paused")

	res, _, err := nobleValidator.ExecQuery(ctx, tokenfactoryModName, "show-minter-controller", roles.MinterController2.FormattedAddress(), "-o", "json")
	require.NoError(t, err, "failed to query minter controller")

	var minterControllerType types.QueryGetMinterControllerResponse
	json.Unmarshal(res, &minterControllerType)

	// minter controller should have been updated even while paused
	minterController2Address := roles.MinterController2.FormattedAddress()
	require.Equal(t, minterController2Address, minterControllerType.MinterController.Controller)

	// minter should have been updated even while paused
	userAddress := extraWallets.User.FormattedAddress()
	require.Equal(t, userAddress, minterControllerType.MinterController.Minter)

	_, err = nobleValidator.ExecTx(ctx, roles.MinterController2.KeyName(),
		tokenfactoryModName, "remove-minter", extraWallets.User.FormattedAddress(),
	)
	require.NoError(t, err, "minters should be able to be removed while in paused state")

	_, err = nobleValidator.ExecTx(ctx, extraWallets.User2.KeyName(), "authz", "grant", extraWallets.Alice.FormattedAddress(), "send", "--spend-limit", fmt.Sprintf("%d%s", 100, mintingDenom))
	require.ErrorContains(t, err, "paused", "failed to grant permissions")

	// IBC transfer fails when asset is paused
	testIBCTransferFail(t, ctx, mintingDenom, noble, gaia, extraWallets.User, extraWallets.User2, "paused")
	testReverseIBCTransferFail(t, ctx, mintingDenom, gaia, noble, gaiaWallets[1], extraWallets.User2, "not found")
	// authz send fails when chain is paused
	testAuthZSendFail(t, ctx, nobleValidator, mintingDenom, noble, extraWallets.User2, extraWallets.User, extraWallets.Alice, "paused", true)
	// authz IBC transfer fails when chain is paused
	testAuthZIBCTransferFail(t, ctx, nobleValidator, mintingDenom, noble, gaia, extraWallets.User2, extraWallets.User, extraWallets.Alice, "paused", true)

	// paying gas in minting denom fails when module is paused
	_, err = nobleValidator.ExecTx(ctx, extraWallets.User.KeyName(),
		"bank", "send", extraWallets.User.KeyName(), extraWallets.Alice.FormattedAddress(), "10token", "--fees", "1"+mintingDenom,
	)
	require.ErrorContains(t, err, "paused", "failed to block spending gas in minting denom")

	_, err = nobleValidator.ExecTx(ctx, roles.Pauser.KeyName(),
		tokenfactoryModName, "unpause",
	)
	require.NoError(t, err, "failed to unpause mints")

	// -- chain unpaused --

	_, err = nobleValidator.ExecTx(ctx, extraWallets.User.KeyName(),
		"bank", "send", extraWallets.User.FormattedAddress(), extraWallets.Alice.FormattedAddress(), "100"+mintingDenom,
	)
	require.NoErrorf(t, err, "failed to send tx bank from user (%s) to alice (%s)", extraWallets.User.FormattedAddress(), extraWallets.Alice.FormattedAddress())

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, math.NewInt(100), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err = noble.GetBalance(ctx, extraWallets.Alice.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get alice balance")
	require.Equal(t, math.NewInt(100), aliceBalance, "alice balance should not have increased while chain is paused")

	testAuthZSendSucceed(t, ctx, nobleValidator, mintingDenom, noble, extraWallets.User, extraWallets.User2, extraWallets.Alice)
	testAuthZIBCTransferSucceed(t, ctx, nobleValidator, mintingDenom, noble, gaia, extraWallets.User2, extraWallets.User, extraWallets.Alice)
}

func testAuthZSend(t *testing.T, ctx context.Context, nobleValidator *cosmos.ChainNode, mintingDenom string, noble *cosmos.CosmosChain, fromWallet ibc.Wallet, toWallet ibc.Wallet, granteeWallet ibc.Wallet, skipGrant bool) (string, error) {
	if !skipGrant {
		_, err := nobleValidator.ExecTx(ctx, fromWallet.KeyName(), "authz", "grant", granteeWallet.FormattedAddress(), "send", "--spend-limit", fmt.Sprintf("%d%s", 100, mintingDenom))
		require.NoError(t, err, "failed to grant permissions")
	}

	bz, _, _ := nobleValidator.ExecBin(ctx, "tx", "bank", "send", fromWallet.FormattedAddress(), toWallet.FormattedAddress(), fmt.Sprintf("%d%s", 50, mintingDenom), "--chain-id", noble.Config().ChainID, "--generate-only")
	_ = nobleValidator.WriteFile(ctx, bz, "tx.json")

	return nobleValidator.ExecTx(ctx, granteeWallet.KeyName(), "authz", "exec", "/var/cosmos-chain/noble-1/tx.json")
}

func testAuthZSendFail(t *testing.T, ctx context.Context, nobleValidator *cosmos.ChainNode, mintingDenom string, noble *cosmos.CosmosChain, fromWallet ibc.Wallet, toWallet ibc.Wallet, granteeWallet ibc.Wallet, errMsg string, skipGrant bool) {
	toWalletInitialBalance := getBalance(t, ctx, mintingDenom, noble, toWallet)

	_, err := testAuthZSend(t, ctx, nobleValidator, mintingDenom, noble, fromWallet, toWallet, granteeWallet, skipGrant)

	require.ErrorContains(t, err, errMsg, "failed to block transactions")
	toWalletBalance := getBalance(t, ctx, mintingDenom, noble, toWallet)
	require.Equal(t, toWalletInitialBalance, toWalletBalance, "toWallet balance should not have incremented")
}

func testAuthZSendSucceed(t *testing.T, ctx context.Context, nobleValidator *cosmos.ChainNode, mintingDenom string, noble *cosmos.CosmosChain, fromWallet ibc.Wallet, toWallet ibc.Wallet, granteeWallet ibc.Wallet) {
	toWalletInitialBalance := getBalance(t, ctx, mintingDenom, noble, toWallet)

	_, err := testAuthZSend(t, ctx, nobleValidator, mintingDenom, noble, fromWallet, toWallet, granteeWallet, false)

	require.NoError(t, err, "failed to send authz transactions")
	toWalletBalance := getBalance(t, ctx, mintingDenom, noble, toWallet)
	require.Equal(t, toWalletInitialBalance+50, toWalletBalance, "toWallet balance should have incremented")
}

func testAuthZIBCTransfer(t *testing.T, ctx context.Context, nobleValidator *cosmos.ChainNode, noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, mintingDenom string, fromWallet ibc.Wallet, toWallet ibc.Wallet, granteeWallet ibc.Wallet, skipGrant bool) (string, error) {
	if !skipGrant {
		_, err := nobleValidator.ExecTx(ctx, fromWallet.KeyName(), "authz", "grant", granteeWallet.FormattedAddress(), "generic", "--msg-type", "/ibc.applications.transfer.v1.MsgTransfer")
		require.NoError(t, err, "failed to grant permissions")
	}

	recipient, err := sdk.Bech32ifyAddressBytes(gaia.Config().Bech32Prefix, toWallet.Address())
	require.NoError(t, err, "failed to convert noble address to gaia address")

	bz, _, _ := nobleValidator.ExecBin(ctx, "tx", "ibc-transfer", "transfer", "transfer", "channel-0", recipient, fmt.Sprintf("%d%s", 50, mintingDenom), "--chain-id", noble.Config().ChainID, "--from", fromWallet.FormattedAddress(), "--generate-only", "--node", fmt.Sprintf("tcp://%s:26657", nobleValidator.HostName()))
	_ = nobleValidator.WriteFile(ctx, bz, "tx.json")

	return nobleValidator.ExecTx(ctx, granteeWallet.KeyName(), "authz", "exec", "/var/cosmos-chain/noble-1/tx.json")
}

func testAuthZIBCTransferFail(t *testing.T, ctx context.Context, nobleValidator *cosmos.ChainNode, mintingDenom string, noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, fromWallet ibc.Wallet, toWallet ibc.Wallet, granteeWallet ibc.Wallet, errMsg string, skipGrant bool) {
	ibcDenom := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: mintingDenom,
	}.IBCDenom()

	fromWalletInitialBalance := getBalance(t, ctx, mintingDenom, noble, fromWallet)
	toWalletInitialBalance := getBalance(t, ctx, ibcDenom, gaia, toWallet)

	_, err := testAuthZIBCTransfer(t, ctx, nobleValidator, noble, gaia, mintingDenom, fromWallet, toWallet, granteeWallet, skipGrant)

	require.ErrorContains(t, err, errMsg)
	fromWalletBalance := getBalance(t, ctx, mintingDenom, noble, fromWallet)
	toWalletBalance := getBalance(t, ctx, ibcDenom, gaia, toWallet)
	require.Equal(t, fromWalletInitialBalance, fromWalletBalance, "fromWallet balance should not have decremented")
	require.Equal(t, toWalletInitialBalance, toWalletBalance, "toWallet balance should not have incremented")
}

func testAuthZIBCTransferSucceed(t *testing.T, ctx context.Context, nobleValidator *cosmos.ChainNode, mintingDenom string, noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, fromWallet ibc.Wallet, toWallet ibc.Wallet, granteeWallet ibc.Wallet) {
	ibcDenom := getIBCDenom(mintingDenom)

	fromWalletInitialBalance := getBalance(t, ctx, mintingDenom, noble, fromWallet)
	toWalletInitialBalance := getBalance(t, ctx, ibcDenom, gaia, toWallet)

	_, err := testAuthZIBCTransfer(t, ctx, nobleValidator, noble, gaia, mintingDenom, fromWallet, toWallet, granteeWallet, false)
	require.NoError(t, err, "failed to exec IBC transfer via authz")

	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	fromWalletBalance := getBalance(t, ctx, mintingDenom, noble, fromWallet)
	require.Equal(t, fromWalletInitialBalance-50, fromWalletBalance, "fromWallet balance should have decremented")
	toWalletBalance := getBalance(t, ctx, ibcDenom, gaia, toWallet)
	require.Equal(t, toWalletInitialBalance+50, toWalletBalance, "toWallet balance should have incremented")
}

func testIBCTransfer(t *testing.T, ctx context.Context, mintingDenom string, noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, fromWallet ibc.Wallet, toWallet ibc.Wallet) (string, error) {
	recipient, err := sdk.Bech32ifyAddressBytes(gaia.Config().Bech32Prefix, toWallet.Address())
	require.NoError(t, err, "failed to convert address")

	validator := noble.Validators[0]

	return validator.ExecTx(ctx, fromWallet.KeyName(), "ibc-transfer", "transfer", "transfer", "channel-0", recipient, fmt.Sprintf("%d%s", 50, mintingDenom), "--chain-id", noble.Config().ChainID, "--from", fromWallet.FormattedAddress(), "--node", fmt.Sprintf("tcp://%s:26657", validator.HostName()))
}

func testIBCTransferFail(t *testing.T, ctx context.Context, mintingDenom string, noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, fromWallet ibc.Wallet, toWallet ibc.Wallet, errMsg string) {
	ibcDenom := getIBCDenom(mintingDenom)
	fromWalletInitialBalance := getBalance(t, ctx, mintingDenom, noble, fromWallet)
	toWalletInitialBalance := getBalance(t, ctx, ibcDenom, gaia, toWallet)

	_, err := testIBCTransfer(t, ctx, mintingDenom, noble, gaia, fromWallet, toWallet)

	require.ErrorContains(t, err, errMsg, "failed to block IBC transfer")
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))
	fromWalletBalance := getBalance(t, ctx, mintingDenom, noble, fromWallet)
	require.Equal(t, fromWalletInitialBalance, fromWalletBalance, "fromWallet balance should not have decremented")
	toWalletBalance := getBalance(t, ctx, ibcDenom, gaia, toWallet)
	require.Equal(t, toWalletInitialBalance, toWalletBalance, "toWallet balance should not have incremented")
}

func testIBCTransferSucceed(t *testing.T, ctx context.Context, mintingDenom string, noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, fromWallet ibc.Wallet, toWallet ibc.Wallet) {
	ibcDenom := getIBCDenom(mintingDenom)
	fromWalletInitialBalance := getBalance(t, ctx, mintingDenom, noble, fromWallet)
	toWalletInitialBalance := getBalance(t, ctx, ibcDenom, gaia, toWallet)

	_, err := testIBCTransfer(t, ctx, mintingDenom, noble, gaia, fromWallet, toWallet)

	require.NoError(t, err, "failed to send IBC transfer")
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))
	fromWalletBalance := getBalance(t, ctx, mintingDenom, noble, fromWallet)
	require.Equal(t, fromWalletInitialBalance-50, fromWalletBalance, "fromWallet balance should have decremented")
	toWalletBalance := getBalance(t, ctx, ibcDenom, gaia, toWallet)
	require.Equal(t, toWalletInitialBalance+50, toWalletBalance, "toWallet balance should have incremented")
}

func testReverseIBCTransferFail(t *testing.T, ctx context.Context, mintingDenom string, gaia *cosmos.CosmosChain, noble *cosmos.CosmosChain, fromWallet ibc.Wallet, toWallet ibc.Wallet, errMsg string) {
	height, err := gaia.Height(ctx)
	require.NoError(t, err, "failed to get noble height")

	userBalBefore := getBalance(t, ctx, mintingDenom, noble, toWallet)

	recipient, err := sdk.Bech32ifyAddressBytes(noble.Config().Bech32Prefix, toWallet.Address())
	require.NoError(t, err, "failed to convert address")
	tx, err := gaia.SendIBCTransfer(ctx, "channel-0", fromWallet.KeyName(), ibc.WalletAmount{
		Address: recipient,
		Denom:   getIBCDenom(mintingDenom),
		Amount:  math.NewInt(10),
	}, ibc.TransferOptions{})
	require.NoError(t, err, "failed to send ibc transfer")

	_, err = testutil.PollForAck(ctx, noble, height, height+10, tx.Packet)
	require.ErrorContains(t, err, errMsg, "Expect ack not found from noble")

	userBalAfter := getBalance(t, ctx, mintingDenom, noble, toWallet)
	require.Equal(t, userBalBefore, userBalAfter, "User wallet balance should not have increased")
}

func getBalance(t *testing.T, ctx context.Context, denom string, chain *cosmos.CosmosChain, wallet ibc.Wallet) int64 {
	addr, err := sdk.Bech32ifyAddressBytes(chain.Config().Bech32Prefix, wallet.Address())
	require.NoError(t, err, "failed to convert address")

	bal, err := chain.GetBalance(ctx, addr, denom)
	require.NoError(t, err, "failed to get user balance")
	return bal.Int64()
}

func getIBCDenom(mintingDenom string) string {
	return transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: mintingDenom,
	}.IBCDenom()
}
