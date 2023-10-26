package simapp

import (
	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcante "github.com/cosmos/ibc-go/v4/modules/core/ante"
	"github.com/cosmos/ibc-go/v4/modules/core/keeper"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC keeper.
type HandlerOptions struct {
	ante.HandlerOptions

	IBCKeeper *keeper.Keeper

	// FiatTokenFactory
	FiatTokenFactoryKeeper *fiattokenfactorykeeper.Keeper
}

type IsPausedDecorator struct {
	fiatTokenFactory *fiattokenfactorykeeper.Keeper
}

func NewIsPausedDecorator(ctf *fiattokenfactorykeeper.Keeper) IsPausedDecorator {
	return IsPausedDecorator{
		fiatTokenFactory: ctf,
	}
}

func (ad IsPausedDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		switch m := m.(type) {
		case *banktypes.MsgSend, *banktypes.MsgMultiSend, *transfertypes.MsgTransfer:
			switch m := m.(type) {
			case *banktypes.MsgSend:
				for _, c := range m.Amount {
					paused, err := checkPausedStatebyTokenFactory(ctx, c, ad.fiatTokenFactory)
					if paused {
						return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
					}
				}
			case *banktypes.MsgMultiSend:
				for _, i := range m.Inputs {
					for _, c := range i.Coins {
						paused, err := checkPausedStatebyTokenFactory(ctx, c, ad.fiatTokenFactory)
						if paused {
							return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
						}
					}
				}
			case *transfertypes.MsgTransfer:
				paused, err := checkPausedStatebyTokenFactory(ctx, m.Token, ad.fiatTokenFactory)
				if paused {
					return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
				}
			default:
				continue
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
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
	for _, m := range msgs {
		switch m := m.(type) {
		case *banktypes.MsgSend, *banktypes.MsgMultiSend, *transfertypes.MsgTransfer:
			switch m := m.(type) {
			case *banktypes.MsgSend:
				for _, c := range m.Amount {
					addresses := []string{m.ToAddress, m.FromAddress}
					blacklisted, address, err := checkForBlacklistedAddressByTokenFactory(ctx, addresses, c, ad.fiattokenfactory)
					if blacklisted {
						return ctx, sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", address)
					}
					if err != nil {
						return ctx, sdkerrors.Wrapf(err, "error decoding address (%s)", address)
					}
				}
			case *banktypes.MsgMultiSend:
				for _, i := range m.Inputs {
					for _, c := range i.Coins {
						addresses := []string{i.Address}
						blacklisted, address, err := checkForBlacklistedAddressByTokenFactory(ctx, addresses, c, ad.fiattokenfactory)
						if blacklisted {
							return ctx, sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", address)
						}
						if err != nil {
							return ctx, sdkerrors.Wrapf(err, "error decoding address (%s)", address)
						}
					}
				}
				for _, o := range m.Outputs {
					for _, c := range o.Coins {
						addresses := []string{o.Address}
						blacklisted, address, err := checkForBlacklistedAddressByTokenFactory(ctx, addresses, c, ad.fiattokenfactory)
						if blacklisted {
							return ctx, sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", address)
						}
						if err != nil {
							return ctx, sdkerrors.Wrapf(err, "error decoding address (%s)", address)
						}
					}
				}
			case *transfertypes.MsgTransfer:
				addresses := []string{m.Sender, m.Receiver}
				blacklisted, address, err := checkForBlacklistedAddressByTokenFactory(ctx, addresses, m.Token, ad.fiattokenfactory)
				if blacklisted {
					return ctx, sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", address)
				}
				if err != nil {
					return ctx, sdkerrors.Wrapf(err, "error decoding address (%s)", address)
				}
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}

// checkForBlacklistedAddressByTokenFactory first checks if the denom being transacted is a mintable asset from a TokenFactory,
// if it is, it checks if the addresses involved in the tx are blacklisted by that specific TokenFactory.
func checkForBlacklistedAddressByTokenFactory(ctx sdk.Context, addresses []string, c sdk.Coin, ctf *fiattokenfactorykeeper.Keeper) (blacklisted bool, blacklistedAddress string, err error) {
	ctfMintingDenom := ctf.GetMintingDenom(ctx)
	if c.Denom == ctfMintingDenom.Denom {
		for _, address := range addresses {
			_, addressBz, err := bech32.DecodeAndConvert(address)
			if err != nil {
				return false, address, err
			}
			_, found := ctf.GetBlacklisted(ctx, addressBz)
			if found {
				return true, address, fiattokenfactorytypes.ErrUnauthorized
			}
		}
	}
	return false, "", nil
}

// NewAnteHandler creates a new ante handler
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.FiatTokenFactoryKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "fiat token factory keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(),
		ante.NewRejectExtensionOptionsDecorator(),
		NewIsBlacklistedDecorator(options.FiatTokenFactoryKeeper),
		NewIsPausedDecorator(options.FiatTokenFactoryKeeper),
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewAnteDecorator(options.IBCKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
