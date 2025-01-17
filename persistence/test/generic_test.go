package test

import (
	"encoding/hex"
	"github.com/pokt-network/pocket/persistence/types"
	"reflect"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/stretchr/testify/require"
)

// TODO(andrew): Be consistent with `GenericParam` and `ActorSpecificParam` throughout the codebase; preferably the latter.

func GetGenericActor[T any](
	protocolActorSchema types.ProtocolActorSchema,
	getActor func(*persistence.PostgresContext, []byte) (T, error),
) func(*persistence.PostgresContext, string) (*types.BaseActor, error) {
	return func(db *persistence.PostgresContext, address string) (*types.BaseActor, error) {
		addr, err := hex.DecodeString(address)
		if err != nil {
			return nil, err
		}
		actor, err := getActor(db, addr)
		if err != nil {
			return nil, err
		}
		baseActor := getActorValues(protocolActorSchema, reflect.Indirect(reflect.ValueOf(actor)))
		return &baseActor, nil
	}
}

func NewTestGenericActor[T any](protocolActorSchema types.ProtocolActorSchema, newActor func() (T, error)) func() (types.BaseActor, error) {
	return func() (types.BaseActor, error) {
		actor, err := newActor()
		if err != nil {
			return types.BaseActor{}, err
		}
		return getActorValues(protocolActorSchema, reflect.Indirect(reflect.ValueOf(actor))), nil
	}
}

func getAllActorsTest[T any](
	t *testing.T,
	db *persistence.PostgresContext,
	getAllActors func(height int64) ([]T, error),
	createTestActor func(*persistence.PostgresContext) (T, error),
	updateActor func(*persistence.PostgresContext, T) error,
	initialCount int,
) {
	// The default test state contains `initialCount` actors
	actors, err := getAllActors(0)
	require.NoError(t, err)
	require.Len(t, actors, initialCount)

	// Add 2 actors at height 1
	db.Height++
	_, err = createTestActor(db)
	require.NoError(t, err)
	_, err = createTestActor(db)
	require.NoError(t, err)

	// Check height 0
	actors, err = getAllActors(0)
	require.NoError(t, err)
	require.Len(t, actors, initialCount)

	// Check height 1
	actors, err = getAllActors(1)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+2)

	// Add 1 actor at height 3
	db.Height++
	db.Height++
	_, err = createTestActor(db)
	require.NoError(t, err)

	// Check height 0
	actors, err = getAllActors(0)
	require.NoError(t, err)
	require.Len(t, actors, initialCount)

	// Check height 1
	actors, err = getAllActors(1)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+2)

	// Check height 2
	actors, err = getAllActors(2)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+2)

	// Check height 3
	actors, err = getAllActors(3)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+3)

	// Update the service nodes at different heights and confirm that count does not change
	for _, actor := range actors {
		db.Height++
		err = updateActor(db, actor)
		require.NoError(t, err)

		// Check that count did not change
		actors, err := getAllActors(db.Height)
		require.NoError(t, err)
		require.Len(t, actors, initialCount+3)
	}

	// Check height 1
	actors, err = getAllActors(1)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+2)

	// Check height 10
	actors, err = getAllActors(10)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+3)
}

func getTestGetSetStakeAmountTest[T any](
	t *testing.T,
	db *persistence.PostgresContext,
	createTestActor func(*persistence.PostgresContext) (*T, error),
	getActorStake func(int64, []byte) (string, error),
	setActorStake func([]byte, string) error,
	height int64,
) {
	var newStakeAmount = "new_stake_amount"

	actor, err := createTestActor(db)
	require.NoError(t, err)
	addrStr := reflect.ValueOf(*actor).FieldByName("Address").String()

	addr, err := hex.DecodeString(addrStr)
	require.NoError(t, err)

	// Check stake amount before
	stakeAmount, err := getActorStake(height, addr)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, stakeAmount, "unexpected beginning stakeAmount")

	// Check stake amount after
	err = setActorStake(addr, newStakeAmount)
	require.NoError(t, err)

	stakeAmountAfter, err := getActorStake(height, addr)
	require.NoError(t, err)
	require.Equal(t, newStakeAmount, stakeAmountAfter, "unexpected status")
}

func getActorValues(_ types.ProtocolActorSchema, actorValue reflect.Value) types.BaseActor {
	chains := make([]string, 0)
	if actorValue.FieldByName("Chains").Kind() != 0 {
		chains = actorValue.FieldByName("Chains").Interface().([]string)
	}

	return types.BaseActor{
		Address:            actorValue.FieldByName("Address").String(),
		PublicKey:          actorValue.FieldByName("PublicKey").String(),
		StakedTokens:       actorValue.FieldByName("StakedAmount").String(),
		ActorSpecificParam: actorValue.FieldByName("GenericParam").String(),
		OutputAddress:      actorValue.FieldByName("Output").String(),
		PausedHeight:       int64(actorValue.FieldByName("PausedHeight").Int()),
		UnstakingHeight:    int64(actorValue.FieldByName("UnstakingHeight").Int()),
		Chains:             chains,
	}
}
