package keeper

import (
	"context"
	"strconv"

	"github.com/alice/checkers/x/checkers/rules"
	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreateGame(goCtx context.Context, msg *types.MsgCreateGame) (*types.MsgCreateGameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	// get the new games id
	systemInfo, found := k.Keeper.GetSystemInfo(ctx)
	if !found {
		panic("SystemInfo not found")
	}
	newIndex := strconv.FormatUint(systemInfo.NextId, 10)
	// create new stored game object
	newGame := rules.New()
	storedGame := types.StoredGame{
		Index:       newIndex,
		Board:       newGame.String(),
		Turn:        rules.PieceStrings[newGame.Turn],
		Black:       msg.Black,
		Red:         msg.Red,
		Winner:      rules.PieceStrings[rules.NO_PLAYER],
		Deadline:    types.FormatDeadline(types.GetNextDeadline(ctx)),
		MoveCount:   0,
		BeforeIndex: types.NoFifoIndex,
		AfterIndex:  types.NoFifoIndex,
		Wager:       msg.Wager,
		Denom:       msg.Denom,
	}
	// confirm addresses
	err := storedGame.Validate()
	if err != nil {
		return nil, err
	}
	//send game to tail
	k.Keeper.SendToFifoTail(ctx, &storedGame, &systemInfo)
	//save stored game object
	k.Keeper.SetStoredGame(ctx, storedGame)
	// set stage for next game
	systemInfo.NextId++
	k.Keeper.SetSystemInfo(ctx, systemInfo)
	// consume gas
	ctx.GasMeter().ConsumeGas(types.CreateGameGas, "Create game")

	//emit event for creating a game
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.GameCreatedEventType,
			sdk.NewAttribute(types.GameCreatedEventCreator, msg.Creator),
			sdk.NewAttribute(types.GameCreatedEventGameIndex, newIndex),
			sdk.NewAttribute(types.GameCreatedEventBlack, msg.Black),
			sdk.NewAttribute(types.GameCreatedEventRed, msg.Red),
			sdk.NewAttribute(types.GameCreatedEventWager, strconv.FormatUint(msg.Wager, 10)),
			sdk.NewAttribute(types.GameCreatedEventDenom, msg.Denom),
		),
	)

	// return new ID for reference
	return &types.MsgCreateGameResponse{
		GameIndex: newIndex,
	}, nil

}
