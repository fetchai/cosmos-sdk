package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	bankKeeper       types.BankKeeper
	cdc              codec.BinaryMarshaler
	storeKey         sdk.StoreKey
	feeCollectorName string
	paramSpace       paramtypes.Subspace
}

type FundPair struct {
	Fund    types.Fund
	Account sdk.AccAddress
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, bankKeeper types.BankKeeper, feeCollectorName string) Keeper {

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		bankKeeper:       bankKeeper,
		cdc:              cdc,
		storeKey:         storeKey,
		feeCollectorName: feeCollectorName,
		paramSpace:       paramSpace,
	}
}

func (k Keeper) AddFund(ctx sdk.Context, sender sdk.AccAddress, fund types.Fund) error {
	params := k.GetParams(ctx)
	if !params.IsAllowedSender(sender) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrConflict,
			"Non-whitelist sender %s", sender.String(),
		)
	}

	// validate that the fund we are adding is correct
	err := fund.ValidateBasic()
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(*fund.Amount))
	if err != nil {
		return err
	}

	return k.setFund(ctx, sender, fund, false);
}

func (k Keeper) UpdateFund(ctx sdk.Context, sender sdk.AccAddress, fund types.Fund) error {
	return k.setFund(ctx, sender, fund, true);
}

func (k Keeper) setFund(ctx sdk.Context, sender sdk.AccAddress, fund types.Fund, shouldExist bool) error {
	key := types.GetActiveFundKey(sender)

	// check that we do not already have fund from this account
	store := ctx.KVStore(k.storeKey)

	// check to see if the fund should exist or not
	if shouldExist {
		if !store.Has(key) {
			return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Fund from sender already exists")
		}

		// if a fund is updated with a zero or negative remaining amount then simple delete the entry
		if fund.Amount.IsNegative() || fund.Amount.IsZero() {
			store.Delete(key) // remove the entry
		} else {
			store.Set(key, k.cdc.MustMarshalBinaryBare(&fund)) // update the entry
		}

	} else {
		if store.Has(key) {
			return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Fund from sender already exists")
		}

		// update the fund
		store.Set(key, k.cdc.MustMarshalBinaryBare(&fund))
	}

	return nil
}

func (k Keeper) GetFund(ctx sdk.Context, sender sdk.AccAddress) (*types.Fund, error) {
	key := types.GetActiveFundKey(sender)

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	if bz == nil {
		return nil, nil
	}

	fund := &types.Fund{}
	k.cdc.MustUnmarshalBinaryBare(bz, fund)
	return fund, nil
}


func (k Keeper) GetAllFunds(ctx sdk.Context) []FundPair {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ActiveFundKeyPrefix)
	defer iter.Close()

	funds := []FundPair{}
	for ; iter.Valid(); iter.Next() {
		pair := FundPair{
			Account: types.GetAddressFromActiveFundKey(iter.Key()),
		}
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &pair.Fund)

		funds = append(funds, pair)
	}

	return funds
}

func (k Keeper) DripAllFunds(ctx sdk.Context) (*sdk.Coins, error) {
	drips := sdk.NewCoins()
	funds := k.GetAllFunds(ctx)

	for _, fund := range funds {
		newFund, drip := fund.Fund.Drip() // calculate the drip for this block

		// update the fund - we either delete or update
		err := k.UpdateFund(ctx, fund.Account, newFund)
		if err != nil {
			continue // ignore this drip
		}

		// update the drips
		drips = drips.Add(drip)
	}

	// send the funds to the fee collector module
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, drips)
	if err != nil {
		return nil, err
	}

	return &drips, nil
}

func (k Keeper) GetActiveFunds(ctx sdk.Context) []types.ActiveFund {
	activeFunds := []types.ActiveFund{}
	for _, fund := range k.GetAllFunds(ctx) {
		activeFunds = append(activeFunds, types.ActiveFund{
			Sender: fund.Account.String(),
			Fund:   &fund.Fund,
		})
	}

	return activeFunds
}

// SetActiveFunds forcibly sets the active funds that should be used
func (k Keeper) SetActiveFunds(ctx sdk.Context, funds []types.ActiveFund) error {
	coins := sdk.NewCoins()

	for _, fund := range funds {
		account, err := sdk.AccAddressFromBech32(fund.Sender)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Invalid address: %s", fund.Sender)
		}

		if fund.Fund == nil {
			return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Invalid fund")
		}

		// update the coins
		coins = coins.Add(*fund.Fund.Amount)

		err = k.setFund(ctx, account, *fund.Fund, false)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Failed to set active fund: %s", err.Error())
		}
	}

	// finally set the balance for this module
	err := k.bankKeeper.SetBalances(ctx, authtypes.NewModuleAddress(types.ModuleName), coins)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Failed to set active coins: %s", err.Error())
	}

	return nil
}