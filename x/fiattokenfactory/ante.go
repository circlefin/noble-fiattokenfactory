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

package fiattokenfactory

import (
	"errors"

	sdkerrors "cosmossdk.io/errors"
	"github.com/btcsuite/btcd/btcutil/bech32"
	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
)

type IsPausedDecorator struct {
	cdc              codec.Codec
	fiatTokenFactory *fiattokenfactorykeeper.Keeper
}

func NewIsPausedDecorator(cdc codec.Codec, ftf *fiattokenfactorykeeper.Keeper) IsPausedDecorator {
	return IsPausedDecorator{
		cdc:              cdc,
		fiatTokenFactory: ftf,
	}
}

func (ad IsPausedDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()

	err = ad.CheckMessages(ctx, msgs)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (ad IsPausedDecorator) CheckMessages(ctx sdk.Context, msgs []sdk.Msg) error {
	for _, msg := range msgs {
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			nestedMsgs, err := execMsg.GetMessages()
			if err != nil {
				return err
			}

			return ad.CheckMessages(ctx, nestedMsgs)
		}

		switch m := msg.(type) {
		case *authz.MsgGrant:
			var authorization authz.Authorization
			err := ad.cdc.UnpackAny(m.Grant.Authorization, &authorization)
			if err != nil {
				return err
			}

			if grant, ok := authorization.(*banktypes.SendAuthorization); ok {
				for _, coin := range grant.SpendLimit {
					paused, err := checkPausedStatebyTokenFactory(ctx, coin, ad.fiatTokenFactory)
					if paused {
						return sdkerrors.Wrapf(err, "can not perform token authorizations")
					}
				}
			}
		default:
			continue
		}
	}

	return nil
}

func checkPausedStatebyTokenFactory(ctx sdk.Context, c sdk.Coin, ctf *fiattokenfactorykeeper.Keeper) (bool, *sdkerrors.Error) {
	ctfMintingDenom := ctf.GetMintingDenom(ctx)
	if c.Denom == ctfMintingDenom.Denom {
		paused := ctf.GetPaused(ctx)
		if paused.Paused {
			return true, fiattokenfactorytypes.ErrPaused
		}
	}
	return false, nil
}

type IsBlacklistedDecorator struct {
	fiattokenfactory *fiattokenfactorykeeper.Keeper
}

func NewIsBlacklistedDecorator(ctf *fiattokenfactorykeeper.Keeper) IsBlacklistedDecorator {
	return IsBlacklistedDecorator{
		fiattokenfactory: ctf,
	}
}

func (ad IsBlacklistedDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()

	err = ad.CheckMessages(ctx, msgs, nil)
	if err != nil {
		return ctx, err
	}

	return next(ad.AddGranteeToContextIfPresent(ctx, msgs), tx, simulate)
}

func (ad IsBlacklistedDecorator) AddGranteeToContextIfPresent(ctx sdk.Context, msgs []sdk.Msg) sdk.Context {
	var grantees []string
	for _, msg := range msgs {
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			grantees = append(grantees, execMsg.Grantee)
		}
	}
	if len(grantees) > 0 {
		return ctx.WithValue(types.GranteeKey, grantees)
	}
	return ctx
}

func (ad IsBlacklistedDecorator) CheckMessages(ctx sdk.Context, msgs []sdk.Msg, grantee *string) error {
	for _, msg := range msgs {
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			nestedMsgs, err := execMsg.GetMessages()
			if err != nil {
				return err
			}

			return ad.CheckMessages(ctx, nestedMsgs, &execMsg.Grantee)
		}

		switch m := msg.(type) {
		case *transfertypes.MsgTransfer:
			// since the Transfer receiver is not on Noble, it is not checked by send restrictions and needs to be checked here
			err := checkForBlacklistedAddressByTokenFactory(ctx, m.Receiver, m.Token, ad.fiattokenfactory)
			if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
				return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not receive tokens", m.Receiver)
			} else if err != nil {
				return sdkerrors.Wrapf(err, "error decoding address (%s)", m.Receiver)
			}
		default:
			continue
		}
	}

	return nil
}

// checkForBlacklistedAddressByTokenFactory first checks if the denom being transacted is a mintable asset from a TokenFactory,
// if it is, it checks if the address involved in the tx is blacklisted by that specific TokenFactory.
func checkForBlacklistedAddressByTokenFactory(ctx sdk.Context, address string, c sdk.Coin, ctf *fiattokenfactorykeeper.Keeper) error {
	ctfMintingDenom := ctf.GetMintingDenom(ctx)
	if c.Denom == ctfMintingDenom.Denom {
		_, addressBz, err := bech32.DecodeToBase256(address)
		if err != nil {
			return err
		}
		_, found := ctf.GetBlacklisted(ctx, addressBz)
		if found {
			return fiattokenfactorytypes.ErrUnauthorized
		}
	}
	return nil
}
