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

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/relayer/rly"
)

var nobleImageInfo = []ibc.DockerImage{
	{
		Repository: "noble-fiattokenfactory-simd",
		Version:    "local",
		UidGid:     "1025:1025",
	},
}

var (
	denomMetadataFrienzies = DenomMetadata{
		Display: "ufrienzies",
		Base:    "ufrienzies",
		Name:    "frienzies",
		Symbol:  "FRNZ",
		DenomUnits: []DenomUnit{
			{
				Denom: "ufrienzies",
				Aliases: []string{
					"microfrienzies",
				},
				Exponent: "0",
			},
			{
				Denom: "mfrienzies",
				Aliases: []string{
					"millifrienzies",
				},
				Exponent: "3",
			},
			{
				Denom:    "frienzies",
				Exponent: "6",
			},
		},
	}

	denomMetadataRupee = DenomMetadata{
		Display: "rupee",
		Base:    "urupee",
		Name:    "rupee",
		Symbol:  "RUPEE",
		DenomUnits: []DenomUnit{
			{
				Denom: "urupee",
				Aliases: []string{
					"microrupee",
				},
				Exponent: "0",
			},
			{
				Denom: "mrupee",
				Aliases: []string{
					"millirupee",
				},
				Exponent: "3",
			},
			{
				Denom:    "rupee",
				Exponent: "6",
			},
		},
	}

	denomMetadataDrachma = DenomMetadata{
		Display: "drachma",
		Base:    "udrachma",
		Name:    "drachma",
		Symbol:  "DRACHMA",
		DenomUnits: []DenomUnit{
			{
				Denom: "udrachma",
				Aliases: []string{
					"microdrachma",
				},
				Exponent: "0",
			},
			{
				Denom: "mdrachma",
				Aliases: []string{
					"millidrachma",
				},
				Exponent: "3",
			},
			{
				Denom:    "drachma",
				Exponent: "6",
			},
		},
	}

	defaultShare                   = "0.8"
	defaultDistributionEntityShare = "1.0"
	defaultTransferBPSFee          = "1"
	defaultTransferMaxFee          = "5000000"
	defaultTransferFeeDenom        = denomMetadataDrachma.Base

	relayerImage = relayer.CustomDockerImage("ghcr.io/cosmos/relayer", "v2.4.2", rly.RlyDefaultUidGid)
)

type DenomMetadata struct {
	Display    string      `json:"display"`
	Base       string      `json:"base"`
	Name       string      `json:"name"`
	Symbol     string      `json:"symbol"`
	DenomUnits []DenomUnit `json:"denom_units"`
}

type DenomUnit struct {
	Denom    string   `json:"denom"`
	Aliases  []string `json:"aliases"`
	Exponent string   `json:"exponent"`
}

type TokenFactoryAddress struct {
	Address string `json:"address"`
}

type ParamAuthAddress struct {
	Address string `json:"address"`
}

type TokenFactoryPaused struct {
	Paused bool `json:"paused"`
}

type TokenFactoryDenom struct {
	Denom string `json:"denom"`
}

type DistributionEntity struct {
	Address string `json:"address"`
	Share   string `json:"share"`
}

type CCTPAmount struct {
	Amount string `json:"amount"`
}

type CCTPPerMessageBurnLimit struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}

type CCTPNumber struct {
	Amount string `json:"amount"`
}

type CCTPNonce struct {
	Nonce string `json:"nonce"`
}

type Attester struct {
	Attester string `json:"attester"`
}

type ExtraWallets struct {
	User  ibc.Wallet
	User2 ibc.Wallet
	Alice ibc.Wallet
}

type NobleRoles struct {
	Owner             ibc.Wallet
	Owner2            ibc.Wallet
	MasterMinter      ibc.Wallet
	MinterController  ibc.Wallet
	MinterController2 ibc.Wallet
	Minter            ibc.Wallet
	Blacklister       ibc.Wallet
	Pauser            ibc.Wallet
}

// Creates tokenfactory wallets. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createTokenfactoryRoles(ctx context.Context, denomMetadata DenomMetadata, val *cosmos.ChainNode, minSetup bool) (NobleRoles, error) {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	var err error

	nobleRoles := NobleRoles{}

	nobleRoles.Owner, err = nobleVal.BuildRelayerWallet(ctx, "owner-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create wallet: %w", err)
	}

	if err := val.RecoverKey(ctx, nobleRoles.Owner.KeyName(), nobleRoles.Owner.Mnemonic()); err != nil {
		return NobleRoles{}, fmt.Errorf("failed to restore %s wallet: %w", nobleRoles.Owner.KeyName(), err)
	}

	genesisWallet := ibc.WalletAmount{
		Address: nobleRoles.Owner.FormattedAddress(),
		Denom:   chainCfg.Denom,
		Amount:  math.ZeroInt(),
	}
	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []types.Coin{types.NewCoin(genesisWallet.Denom, genesisWallet.Amount)})
	if err != nil {
		return NobleRoles{}, err
	}
	if minSetup {
		return nobleRoles, nil
	}

	nobleRoles.Owner2, err = nobleVal.BuildRelayerWallet(ctx, "owner2-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "owner2", err)
	}
	nobleRoles.MasterMinter, err = nobleVal.BuildRelayerWallet(ctx, "masterminter-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "masterminter", err)
	}
	nobleRoles.MinterController, err = nobleVal.BuildRelayerWallet(ctx, "mintercontroller-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "mintercontroller", err)
	}
	nobleRoles.MinterController2, err = nobleVal.BuildRelayerWallet(ctx, "mintercontroller2-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "mintercontroller2", err)
	}
	nobleRoles.Minter, err = nobleVal.BuildRelayerWallet(ctx, "minter-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "minter", err)
	}
	nobleRoles.Blacklister, err = nobleVal.BuildRelayerWallet(ctx, "blacklister-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "blacklister", err)
	}
	nobleRoles.Pauser, err = nobleVal.BuildRelayerWallet(ctx, "pauser-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "pauser", err)
	}

	walletsToRestore := []ibc.Wallet{nobleRoles.Owner2, nobleRoles.MasterMinter, nobleRoles.MinterController, nobleRoles.MinterController2, nobleRoles.Minter, nobleRoles.Blacklister, nobleRoles.Pauser}
	for _, wallet := range walletsToRestore {
		if err = val.RecoverKey(ctx, wallet.KeyName(), wallet.Mnemonic()); err != nil {
			return NobleRoles{}, fmt.Errorf("failed to restore %s wallet: %w", wallet.KeyName(), err)
		}
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: nobleRoles.Owner2.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.MasterMinter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.MinterController.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.MinterController2.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.Minter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.Blacklister.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.Pauser.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []types.Coin{types.NewCoin(wallet.Denom, wallet.Amount)})
		if err != nil {
			return NobleRoles{}, err
		}
	}

	return nobleRoles, nil
}

// Creates extra wallets used for testing. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createParamAuthAtGenesis(ctx context.Context, val *cosmos.ChainNode) (ibc.Wallet, error) {
	chainCfg := val.Chain.Config()

	wallet, err := val.Chain.BuildWallet(ctx, "authority", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	genesisWallet := ibc.WalletAmount{
		Address: wallet.FormattedAddress(),
		Denom:   chainCfg.Denom,
		Amount:  math.ZeroInt(),
	}

	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []types.Coin{types.NewCoin(genesisWallet.Denom, genesisWallet.Amount)})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

// Creates extra wallets used for testing. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createExtraWalletsAtGenesis(ctx context.Context, val *cosmos.ChainNode) (ExtraWallets, error) {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	var err error

	extraWallets := &ExtraWallets{}

	extraWallets.User, err = nobleVal.BuildRelayerWallet(ctx, "user")
	if err != nil {
		return ExtraWallets{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	extraWallets.User2, err = nobleVal.BuildRelayerWallet(ctx, "user2")
	if err != nil {
		return ExtraWallets{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	extraWallets.Alice, err = nobleVal.BuildRelayerWallet(ctx, "alice")
	if err != nil {
		return ExtraWallets{}, fmt.Errorf("failed to create wallet: %w", err)
	}

	walletsToRestore := []ibc.Wallet{extraWallets.User, extraWallets.User2, extraWallets.Alice}
	for _, wallet := range walletsToRestore {
		if err = val.RecoverKey(ctx, wallet.KeyName(), wallet.Mnemonic()); err != nil {
			return ExtraWallets{}, fmt.Errorf("failed to restore %s wallet: %w", wallet.KeyName(), err)
		}
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: extraWallets.User.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: extraWallets.User2.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.NewInt(10_000),
		},
		{
			Address: extraWallets.Alice.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []types.Coin{types.NewCoin(wallet.Denom, wallet.Amount)})
		if err != nil {
			return ExtraWallets{}, err
		}
	}
	return *extraWallets, nil
}

type genesisWrapper struct {
	chain          *cosmos.CosmosChain
	tfRoles        NobleRoles
	fiatTfRoles    NobleRoles
	paramAuthority ibc.Wallet
	extraWallets   ExtraWallets
}

func nobleChainSpec(
	ctx context.Context,
	gw *genesisWrapper,
	chainID string,
	nv, nf int,
	minSetupFiatTf bool,
	minModifyFiatTf bool,
) *interchaintest.ChainSpec {
	return &interchaintest.ChainSpec{
		NumValidators: &nv,
		NumFullNodes:  &nf,
		ChainConfig: ibc.ChainConfig{
			Type:           "cosmos",
			Name:           "noble",
			ChainID:        chainID,
			Bin:            "simd", // "nobled",
			Denom:          "token",
			Bech32Prefix:   "noble",
			CoinType:       "118",
			GasPrices:      "0.0token",
			GasAdjustment:  1.1,
			TrustingPeriod: "504h",
			NoHostMount:    false,
			Images:         nobleImageInfo,
			PreGenesis:     preGenesisAll(ctx, gw, minSetupFiatTf),
			ModifyGenesis:  modifyGenesisAll(gw, minModifyFiatTf),
		},
	}
}

func preGenesisAll(ctx context.Context, gw *genesisWrapper, minSetupFiatTf bool) func(ibc.ChainConfig) error {
	return func(cc ibc.ChainConfig) (err error) {
		val := gw.chain.Validators[0]

		gw.fiatTfRoles, err = createTokenfactoryRoles(ctx, denomMetadataDrachma, val, minSetupFiatTf)
		if err != nil {
			return err
		}

		gw.extraWallets, err = createExtraWalletsAtGenesis(ctx, val)
		if err != nil {
			return err
		}

		return err
	}
}

func modifyGenesisAll(gw *genesisWrapper, minSetupFiatTf bool) func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
	return func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
		g := make(map[string]interface{})

		if err := json.Unmarshal(b, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}

		if err := modifyGenesisTokenfactory(g, "fiattokenfactory", denomMetadataDrachma, gw.fiatTfRoles, minSetupFiatTf); err != nil {
			return nil, err
		}

		out, err := json.Marshal(&g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}

		return out, nil
	}
}

// Modifies tokenfactory genesis accounts.
// If minSetup = true, only the owner address, paused state, and denom is setup in genesis.
// These are minimum requirements to start the chain. Otherwise all tokenfactory accounts are created.
func modifyGenesisTokenfactory(g map[string]interface{}, tokenfactoryModName string, denomMetadata DenomMetadata, roles NobleRoles, minSetup bool) error {
	if err := dyno.Set(g, TokenFactoryAddress{roles.Owner.FormattedAddress()}, "app_state", tokenfactoryModName, "owner"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryPaused{false}, "app_state", tokenfactoryModName, "paused"); err != nil {
		return fmt.Errorf("failed to set paused in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryDenom{denomMetadata.Base}, "app_state", tokenfactoryModName, "mintingDenom"); err != nil {
		return fmt.Errorf("failed to set minting denom in genesis json: %w", err)
	}
	if err := dyno.Append(g, denomMetadata, "app_state", "bank", "denom_metadata"); err != nil {
		return fmt.Errorf("failed to set denom metadata in genesis json: %w", err)
	}
	if minSetup {
		return nil
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.MasterMinter.FormattedAddress()}, "app_state", tokenfactoryModName, "masterMinter"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Blacklister.FormattedAddress()}, "app_state", tokenfactoryModName, "blacklister"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Pauser.FormattedAddress()}, "app_state", tokenfactoryModName, "pauser"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	return nil
}
