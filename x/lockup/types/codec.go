package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLockTokens{}, "osmosis/lockup/lock-tokens", nil)
	cdc.RegisterConcrete(&MsgBeginUnlockingAll{}, "osmosis/lockup/begin-unlock-tokens", nil)
	cdc.RegisterConcrete(&MsgBeginUnlocking{}, "osmosis/lockup/begin-unlock-period-lock", nil)
	cdc.RegisterConcrete(&MsgExtendLockup{}, "osmosis/lockup/extend-lockup", nil)
	cdc.RegisterConcrete(&MsgForceUnlock{}, "osmosis/lockup/force-unlock-tokens", nil)
	cdc.RegisterConcrete(&MsgSetRewardReceiverAddress{}, "osmosis/lockup/set-reward-receiver-address", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgLockTokens{},
		&MsgBeginUnlockingAll{},
		&MsgBeginUnlocking{},
		&MsgExtendLockup{},
		&MsgForceUnlock{},
		&MsgSetRewardReceiverAddress{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
