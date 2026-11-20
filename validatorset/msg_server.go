package validatorset

import "context"

type msgServer struct{ keeper *Keeper }

var _ MsgServer = msgServer{}

func NewMsgServer(keeper *Keeper) MsgServer { return msgServer{keeper: keeper} }

func (server msgServer) AddValidator(context.Context, *MsgAddValidator) (*MsgAddValidatorResponse, error) {
	return nil, ErrUnauthorized
}
func (server msgServer) SetValidatorPower(context.Context, *MsgSetValidatorPower) (*MsgSetValidatorPowerResponse, error) {
	return nil, ErrUnauthorized
}
func (server msgServer) RemoveValidator(context.Context, *MsgRemoveValidator) (*MsgRemoveValidatorResponse, error) {
	return nil, ErrUnauthorized
}
