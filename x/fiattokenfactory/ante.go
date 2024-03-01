package fiattokenfactory

import (
	"errors"

	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
)

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
		case *banktypes.MsgSend:
			for _, c := range m.Amount {
				paused, err := checkPausedStatebyTokenFactory(ctx, c, ad.fiatTokenFactory)
				if paused {
					return sdkerrors.Wrapf(err, "can not perform token transfers")
				}
			}
		case *banktypes.MsgMultiSend:
			for _, i := range m.Inputs {
				for _, c := range i.Coins {
					paused, err := checkPausedStatebyTokenFactory(ctx, c, ad.fiatTokenFactory)
					if paused {
						return sdkerrors.Wrapf(err, "can not perform token transfers")
					}
				}
			}
		case *transfertypes.MsgTransfer:
			paused, err := checkPausedStatebyTokenFactory(ctx, m.Token, ad.fiatTokenFactory)
			if paused {
				return sdkerrors.Wrapf(err, "can not perform token transfers")
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

	return next(ctx, tx, simulate)
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
		case *banktypes.MsgSend:
			for _, c := range m.Amount {
				if grantee != nil {
					err := checkForBlacklistedAddressByTokenFactory(ctx, *grantee, c, ad.fiattokenfactory)
					if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
						return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not receive tokens", *grantee)
					} else if err != nil {
						return sdkerrors.Wrapf(err, "error decoding address (%s)", *grantee)
					}
				}

				err := checkForBlacklistedAddressByTokenFactory(ctx, m.ToAddress, c, ad.fiattokenfactory)
				if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
					return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not receive tokens", m.ToAddress)
				} else if err != nil {
					return sdkerrors.Wrapf(err, "error decoding address (%s)", m.ToAddress)
				}
				err = checkForBlacklistedAddressByTokenFactory(ctx, m.FromAddress, c, ad.fiattokenfactory)
				if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
					return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send tokens", m.FromAddress)
				} else if err != nil {
					return sdkerrors.Wrapf(err, "error decoding address (%s)", m.FromAddress)
				}
			}
		case *banktypes.MsgMultiSend:
			for _, i := range m.Inputs {
				for _, c := range i.Coins {
					if grantee != nil {
						err := checkForBlacklistedAddressByTokenFactory(ctx, *grantee, c, ad.fiattokenfactory)
						if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
							return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not receive tokens", *grantee)
						} else if err != nil {
							return sdkerrors.Wrapf(err, "error decoding address (%s)", *grantee)
						}
					}

					err := checkForBlacklistedAddressByTokenFactory(ctx, i.Address, c, ad.fiattokenfactory)
					if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
						return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", i.Address)
					} else if err != nil {
						return sdkerrors.Wrapf(err, "error decoding address (%s)", i.Address)
					}
				}
			}
			for _, o := range m.Outputs {
				for _, c := range o.Coins {
					if grantee != nil {
						err := checkForBlacklistedAddressByTokenFactory(ctx, *grantee, c, ad.fiattokenfactory)
						if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
							return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not receive tokens", *grantee)
						} else if err != nil {
							return sdkerrors.Wrapf(err, "error decoding address (%s)", *grantee)
						}
					}

					err := checkForBlacklistedAddressByTokenFactory(ctx, o.Address, c, ad.fiattokenfactory)
					if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
						return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", o.Address)
					} else if err != nil {
						return sdkerrors.Wrapf(err, "error decoding address (%s)", o.Address)
					}
				}
			}
		case *transfertypes.MsgTransfer:
			if grantee != nil {
				err := checkForBlacklistedAddressByTokenFactory(ctx, *grantee, m.Token, ad.fiattokenfactory)
				if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
					return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not receive tokens", *grantee)
				} else if err != nil {
					return sdkerrors.Wrapf(err, "error decoding address (%s)", *grantee)
				}
			}

			err := checkForBlacklistedAddressByTokenFactory(ctx, m.Sender, m.Token, ad.fiattokenfactory)
			if errors.Is(err, fiattokenfactorytypes.ErrUnauthorized) {
				return sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send tokens", m.Sender)
			} else if err != nil {
				return sdkerrors.Wrapf(err, "error decoding address (%s)", m.Sender)
			}
			err = checkForBlacklistedAddressByTokenFactory(ctx, m.Receiver, m.Token, ad.fiattokenfactory)
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
		_, addressBz, err := bech32.DecodeAndConvert(address)
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
