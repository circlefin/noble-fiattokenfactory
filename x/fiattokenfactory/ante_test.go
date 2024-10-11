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

package fiattokenfactory_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"
)

var (
	uusdcCoin                  = sdk.NewInt64Coin("uusdc", 10)
	uusdcCoins                 = sdk.Coins{uusdcCoin}
	testAccount1, testAccount2 = sample.TestAccount(), sample.TestAccount()
	TestAccountBech32m         = sample.TestAccountBech32m()
)

func TestAnteHandlerIsPaused(t *testing.T) {
	// ARRANGE: Arrange table driven test cases
	testCases := map[string]struct {
		expectedFailOnPause bool
		message             sdk.Msg
	}{
		"no message": {
			expectedFailOnPause: false,
		},
		"irrelevant msg": {
			expectedFailOnPause: false,
			message:             &testdata.MsgCreateDog{},
		},
		"msgGrant": {
			expectedFailOnPause: true,
			message: func() sdk.Msg {
				mockTime := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)
				mockExpires := mockTime.Add(time.Hour)
				sendAuthz := banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewCoin("uusdc", math.NewInt(1))), nil)
				sendGrant, err := authz.NewGrant(mockTime, sendAuthz, &mockExpires)
				require.NoError(t, err)

				msg := &authz.MsgGrant{
					Granter: "mock",
					Grantee: "mock",
					Grant:   sendGrant,
				}
				return msg
			}(),
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// ARRANGE: setup tokenfactory and isPaused decorator
			ftf, ctx := keeper.FiatTokenfactoryKeeper()
			ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
			ftf.SetPaused(ctx, types.Paused{Paused: false})
			cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
			ad := fiattokenfactory.NewIsPausedDecorator(cdc, ftf)

			// ARRANGE: Build transactions with specific test case message
			builder, err := newMockTxBuilder(cdc)
			require.NoError(t, err)
			if tc.message != nil {
				err = builder.SetMsgs(tc.message)
				require.NoError(t, err)
			}
			tx := builder.GetTx()

			// ACT: Run transaction through ante handler while chain is NOT paused.
			_, err = ad.AnteHandle(ctx, tx, true, mockNext)

			// ASSERT: No errors while chain is not paused
			require.NoError(t, err)

			// ARRANGE: Pause tokenfactory
			ftf.SetPaused(ctx, types.Paused{Paused: true})

			// ACT: Run transaction through ante handler while chain IS paused.
			_, err = ad.AnteHandle(ctx, tx, true, mockNext)

			// ASSERT: Assert expected isPaused error for specific test case messages
			if tc.expectedFailOnPause {
				require.ErrorIs(t, err, types.ErrPaused)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAnteHandlerIsBlacklisted(t *testing.T) {
	// ARRANGE: Arrange table driven test cases
	testCases := map[string]struct {
		message sdk.Msg
		// if blacklistAddressBz is set, the test case will run the message through the antehandler one additional time:
		// 	- without blacklisting any address
		// 	- and also, blacklisting the specified address bytes
		blacklistAddressBz []byte
		// set testInvalidAddress to true if testing for an invalid address or an address that
		// cannot be bech32 decoded.
		testInvalidAddress bool
		expectedError      error
	}{
		"no message": {},
		"irrelevant msg": {
			message: &testdata.MsgCreateDog{},
		},
		"msgTransfer blocked receiver": {
			message: &transfertypes.MsgTransfer{
				Sender:   testAccount1.Address,
				Receiver: testAccount2.Address,
				Token:    uusdcCoin,
			},
			blacklistAddressBz: testAccount2.AddressBz,
			expectedError:      types.ErrUnauthorized,
		},
		"msgTransfer blocked bech32m receiver": {
			message: &transfertypes.MsgTransfer{
				Sender:   testAccount1.Address,
				Receiver: TestAccountBech32m.Address,
				Token:    uusdcCoin,
			},
			blacklistAddressBz: TestAccountBech32m.AddressBz,
			expectedError:      types.ErrUnauthorized,
		},
		"msgTransfer invalid receiver": {
			message: &transfertypes.MsgTransfer{
				Sender:   testAccount1.Address,
				Receiver: "invalid address",
				Token:    uusdcCoin,
			},
			testInvalidAddress: true,
			expectedError:      bech32.ErrInvalidCharacter(32),
		},
		"msgExec MsgTransfer": {
			message: func() sdk.Msg {
				msgTransfer := &transfertypes.MsgTransfer{
					Sender:   testAccount2.Address,
					Receiver: testAccount2.Address,
					Token:    uusdcCoin,
				}
				msgSendAny, err := codectypes.NewAnyWithValue(msgTransfer)
				require.NoError(t, err)
				msg := &authz.MsgExec{
					Grantee: testAccount1.Address,
					Msgs:    []*codectypes.Any{msgSendAny},
				}
				return msg
			}(),
			blacklistAddressBz: testAccount2.AddressBz,
			expectedError:      types.ErrUnauthorized,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// ARRANGE: setup tokenfactory and isBlacklisted decorator
			ftf, ctx := keeper.FiatTokenfactoryKeeper()
			ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
			ftf.SetPaused(ctx, types.Paused{Paused: false})
			cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
			ad := fiattokenfactory.NewIsBlacklistedDecorator(ftf)

			// ARRANGE: Build transactions with specific test case message
			builder, err := newMockTxBuilder(cdc)
			require.NoError(t, err)
			if tc.message != nil {
				err = builder.SetMsgs(tc.message)
				require.NoError(t, err)
			}
			tx := builder.GetTx()

			// ACT: Run transaction through ante handler without blacklisting
			_, err = ad.AnteHandle(ctx, tx, true, mockNext)

			// ASSERT: If we are testing for an invalid address, raise error here
			if tc.testInvalidAddress {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tc.blacklistAddressBz != nil {
				// ARRANGE: Blacklist account
				ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: tc.blacklistAddressBz})

				// ACT: Run transaction through ante handler while account is blacklisted
				_, err = ad.AnteHandle(ctx, tx, true, mockNext)

				// ASSERT: Assert that the unauthorized error is raised
				require.ErrorIs(t, err, tc.expectedError)

				// ARRANGE: Un-blacklist account
				ftf.RemoveBlacklisted(ctx, tc.blacklistAddressBz)
			}
		})
	}
}

func TestAddGranteeToContextIfPresent(t *testing.T) {
	// ARRANGE: Arrange table driven test cases
	testCases := map[string]struct {
		messages         []sdk.Msg
		expectedGrantees []string
	}{
		"no grantees": {
			messages: []sdk.Msg{
				&transfertypes.MsgTransfer{
					Sender:   testAccount1.Address,
					Receiver: testAccount2.Address,
					Token:    uusdcCoin,
				},
				&banktypes.MsgSend{
					FromAddress: testAccount1.Address,
					ToAddress:   testAccount2.Address,
					Amount:      uusdcCoins,
				},
			},
			expectedGrantees: nil,
		},
		"one grantee": {
			messages: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: testAccount1.Address,
					ToAddress:   testAccount2.Address,
					Amount:      uusdcCoins,
				},
				constructMsgExec(t, testAccount1.Address),
			},
			expectedGrantees: []string{testAccount1.Address},
		},
		"multiple grantees": {
			messages: []sdk.Msg{
				constructMsgExec(t, testAccount1.Address),
				constructMsgExec(t, testAccount2.Address),
			},
			expectedGrantees: []string{testAccount1.Address, testAccount2.Address},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// ARRANGE: setup tokenfactory and isBlacklisted decorator
			ftf, ctx := keeper.FiatTokenfactoryKeeper()
			ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"})
			ftf.SetPaused(ctx, types.Paused{Paused: false})
			ad := fiattokenfactory.NewIsBlacklistedDecorator(ftf)

			// ACT: Run transaction through ante handler without blacklisting
			updatedCtx := ad.AddGranteeToContextIfPresent(ctx, tc.messages)

			// ASSERT: Compare the updated context grantees to the expected grantees
			grantees := updatedCtx.Value(types.GranteeKey)
			if tc.expectedGrantees == nil {
				require.Nil(t, grantees)
			} else {
				require.NotNil(t, grantees)
				require.ElementsMatch(t, grantees, tc.expectedGrantees)
			}
		})
	}
}

func constructMsgExec(t *testing.T, granteeAddress string) sdk.Msg {
	mgsSend := &banktypes.MsgSend{
		FromAddress: testAccount2.Address,
		ToAddress:   testAccount2.Address,
		Amount:      uusdcCoins,
	}
	msgSendAny, err := codectypes.NewAnyWithValue(mgsSend)
	require.NoError(t, err)
	msg := &authz.MsgExec{
		Grantee: granteeAddress,
		Msgs:    []*codectypes.Any{msgSendAny},
	}
	return msg
}

func mockNext(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
	return ctx, nil
}

func newMockTxBuilder(cdc codec.ProtoCodecMarshaler) (client.TxBuilder, error) {
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	builder := txConfig.NewTxBuilder()
	privKey := secp256k1.GenPrivKeyFromSecret([]byte("test"))
	pubKey := privKey.PubKey()
	return builder, builder.SetSignatures(
		signingtypes.SignatureV2{
			PubKey:   pubKey,
			Sequence: 0,
			Data:     &signingtypes.SingleSignatureData{},
		},
	)
}
