package simulation_test

import (
	"math/rand"
	"testing"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestProposalMsgs(t *testing.T) {
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	ctx := sdk.NewContext(nil, tmproto.Header{}, true, nil)
	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs()
	assert.Assert(t, len(weightedProposalMsgs) == 1)

	w0 := weightedProposalMsgs[0]

	// tests w0 interface:
	assert.Equal(t, simulation.OpWeightMsgUpdateParams, w0.AppParamsKey())
	assert.Equal(t, simulation.DefaultWeightMsgUpdateParams, w0.DefaultWeight())

	msg := w0.MsgSimulatorFn()(r, ctx, accounts)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	assert.Assert(t, ok)

	assert.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgUpdateParams.Authority)
	assert.DeepEqual(t, sdk.NewDec(0), msgUpdateParams.Params.CommunityTax)
	assert.Equal(t, true, msgUpdateParams.Params.WithdrawAddrEnabled)
}

func TestProposalContents(t *testing.T) {
	var distrKeeper keeper.Keeper
	app, err := simtestutil.Setup(testutil.AppConfig, &distrKeeper)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(distrKeeper)
	require.Len(t, weightedProposalContent, 1)

	w0 := weightedProposalContent[0]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightSubmitCommunitySpendProposal, w0.AppParamsKey())
	require.Equal(t, simtestutil.DefaultWeightTextProposal, w0.DefaultWeight())

	amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)), sdk.NewCoin("atoken", sdk.NewInt(2)))

	feePool := distrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(amount...)
	distrKeeper.SetFeePool(ctx, feePool)

	content := w0.ContentSimulatorFn()(r, ctx, accounts)

	require.Equal(t, "sTxPjfweXhSUkMhPjMaxKlMIJMOXcnQfyzeOcbWwNbeHVIkPZBSpYuLyYggwexjxusrBqDOTtGTOWeLrQKjLxzIivHSlcxgdXhhu", content.GetDescription())
	require.Equal(t, "xKGLwQvuyN", content.GetTitle())
	require.Equal(t, "distribution", content.ProposalRoute())
	require.Equal(t, "CommunityPoolSpend", content.ProposalType())
}
