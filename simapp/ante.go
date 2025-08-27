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

package simapp

import (
	"cosmossdk.io/errors"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory"
	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcante "github.com/cosmos/ibc-go/v10/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
)

var ErrLogic = errors.Register("simapp", 1, "internal logic error")

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC keeper.
type HandlerOptions struct {
	ante.HandlerOptions

	IBCKeeper *ibckeeper.Keeper

	// FiatTokenFactory
	cdc                    codec.Codec
	FiatTokenFactoryKeeper *fiattokenfactorykeeper.Keeper
}

// NewAnteHandler creates a new ante handler
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errors.Wrap(ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, errors.Wrap(ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.FiatTokenFactoryKeeper == nil {
		return nil, errors.Wrap(ErrLogic, "fiat token factory keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, errors.Wrap(ErrLogic, "sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		fiattokenfactory.NewIsBlacklistedDecorator(options.FiatTokenFactoryKeeper),
		fiattokenfactory.NewIsPausedDecorator(options.cdc, options.FiatTokenFactoryKeeper),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
