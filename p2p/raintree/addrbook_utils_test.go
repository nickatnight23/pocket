package raintree

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/pokt-network/pocket/p2p/types"
	"github.com/stretchr/testify/assert"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

type ExpectedRainTreeNetworkConfig struct {
	numNodes          int
	numExpectedLevels int
}

type ExpectedRainTreeMessageTarget struct {
	level int
	left  string
	right string
}
type ExpectedRainTreeMessageProp struct {
	orig     byte
	numNodes int
	addrList string
	targets  []ExpectedRainTreeMessageTarget
}

// IMPROVE(team): Looking into adding more tests and accounting for more edge cases.

func TestRainTreeAddrBookUtilsHandleUpdate(t *testing.T) {
	addr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)

	testCases := []ExpectedRainTreeNetworkConfig{
		// 0 levels
		{1, 0}, // Just self
		// 1 level
		{2, 1},
		{3, 1},
		// 2 levels
		{4, 2},
		{9, 2},
		// 3 levels
		{10, 3},
		{27, 3},
		// 4 levels
		{28, 4},
		{81, 4},
		// 5 levels
		{82, 5},
		// 10 levels
		{59049, 10},
		// 11 levels
		{59050, 11},
		// 19 levels
		// NOTE: This does not scale to 1,000,000,000 (1B) nodes because it's too slow.
		//       However, optimizing the code to handle 1B nodes would be a very premature optimization
		//       at this stage in the project's lifecycle, so the comment is simply left to inform
		//       future readers.
		// {1000000000, 19},
	}

	for _, testCase := range testCases {
		n := testCase.numNodes
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			addrBook := getAddrBook(t, n-1)
			addrBook = append(addrBook, &types.NetworkPeer{Address: addr})
			network := NewRainTreeNetwork(addr, addrBook).(*rainTreeNetwork)

			err = network.processAddrBookUpdates()
			require.NoError(t, err)

			require.Equal(t, len(network.addrList), n)
			require.Equal(t, len(network.addrBookMap), n)
			require.Equal(t, int(network.maxNumLevels), testCase.numExpectedLevels)
		})
	}
}

func BenchmarkAddrBookUpdates(b *testing.B) {
	addr, err := cryptoPocket.GenerateAddress()
	require.NoError(b, err)

	testCases := []ExpectedRainTreeNetworkConfig{
		// Small
		{9, 2},
		// Large
		{59050, 11},
		// INVESTIGATE(olshansky/team): Does not scale to 1,000,000,000 nodes
		// {1000000000, 19},
	}

	for _, testCase := range testCases {
		n := testCase.numNodes
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			addrBook := getAddrBook(nil, n-1)
			addrBook = append(addrBook, &types.NetworkPeer{Address: addr})
			network := NewRainTreeNetwork(addr, addrBook).(*rainTreeNetwork)

			err = network.processAddrBookUpdates()
			require.NoError(b, err)

			require.Equal(b, len(network.addrList), n)
			require.Equal(b, len(network.addrBookMap), n)
			require.Equal(b, int(network.maxNumLevels), testCase.numExpectedLevels)
		})
	}
}

// Generates an address book with a random set of `n` addresses
func getAddrBook(t *testing.T, n int) (addrBook types.AddrBook) {
	addrBook = make([]*types.NetworkPeer, 0)
	for i := 0; i < n; i++ {
		addr, err := cryptoPocket.GenerateAddress()
		if t != nil {
			require.NoError(t, err)
		}
		addrBook = append(addrBook, &types.NetworkPeer{Address: addr})
	}
	return
}

func TestRainTreeAddrBookTargetsSixNodes(t *testing.T) {
	// 		                     A
	// 		   ┌─────────────────┬─────────────────┐
	// 		   C                 A                 E
	//   ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐
	//   D     C     E     B     A     C     F     E     A
	prop := &ExpectedRainTreeMessageProp{'A', 6, "ABCDEF", []ExpectedRainTreeMessageTarget{
		{2, "C", "E"},
		{1, "B", "C"},
		{0, "C", "E"},  // redundancy
		{-1, "F", "B"}, // cleanup
	}}
	testRainTreeMessageTargets(t, prop)
}

func TestRainTreeAddrBookTargetsNineNodes(t *testing.T) {
	//                         A
	//       ┌─────────────────┬─────────────────┐
	//       D                 A                 G
	// ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐
	// F     D     H     C     A     E     I     G     B
	prop := &ExpectedRainTreeMessageProp{'A', 9, "ABCDEFGHI", []ExpectedRainTreeMessageTarget{
		{2, "D", "G"},
		{1, "C", "E"},
		{0, "D", "G"},  // redundancy
		{-1, "I", "B"}, // cleanup
	}}
	testRainTreeMessageTargets(t, prop)
}

func TestRainTreeAddrBookTargetsTwentySevenNodes(t *testing.T) {

	// 		                                                                         O
	// 		                  ┌──────────────────────────────────────────────────────┬─────────────────────────────────────────────────────┐
	// 		                  X                                                      O                                                     F
	//       ┌────────────────┬────────────────┐                    ┌────────────────┬─────────────────┐                 ┌─────────────────┬────────────────────┐
	//       C                X                I                    U                O                 [                 L                 F                    R
	// ┌─────┬─────┐    ┌─────┬─────┐    ┌─────┬─────┐        ┌─────┬─────┐    ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐
	// G     C     K    A     X     E    M     I     Z        Y     U     B    S     O     W     D     [     Q     P     L     T     J     F     N     V     R     H
	prop := &ExpectedRainTreeMessageProp{'O', 27, "OPQRSTUVWXYZ[ABCDEFGHIJKLMN", []ExpectedRainTreeMessageTarget{
		{3, "X", "F"},
		{2, "U", "["},
		{1, "S", "W"},
		{0, "X", "F"}, // redundancy layer
		// 		                                                                         O
		// 		                  ┌──────────────────────────────────────────────────────┬─────────────────────────────────────────────────────┐
		// 		                  X                                                      O                                                     F
		//       ┌────────────────┬────────────────┐                    ┌────────────────┬─────────────────┐                 ┌─────────────────┬────────────────────┐
		//       C                X                I                    U                O                 [                 L                 F                    R
		// ┌─────┬─────┐    ┌─────┬─────┐    ┌─────┬─────┐        ┌─────┬─────┐    ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐
		// G     C     K    A    *X*    E    M     I     Z        Y     U     B    S    *O*    W     D     [     Q     P     L     T     J    *F*    N     V     R     H
		//                        ^                                                      |                                                     ^
		//                        |    <─      <─      <─     <─     <─     <─     <─   <──>    ─>      ─>      ─>      ─>      ─>      ─>     |
		{-1, "N", "P"}, // cleanup  layer
		// 		                                                                         O
		// 		                  ┌──────────────────────────────────────────────────────┬─────────────────────────────────────────────────────┐
		// 		                  X                                                      O                                                     F
		//       ┌────────────────┬────────────────┐                    ┌────────────────┬─────────────────┐                 ┌─────────────────┬────────────────────┐
		//       C                X                I                    U                O                 [                 L                 F                    R
		// ┌─────┬─────┐    ┌─────┬─────┐    ┌─────┬─────┐        ┌─────┬─────┐    ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐     ┌─────┬─────┐
		// G     C     K    A    *X*    E    M     I     Z        Y     U     B   *S* <-*O*-> *W*     D     [     Q     P     L     T     J    *F*    N     V     R     H
	}}
	testRainTreeMessageTargets(t, prop)
}

func testRainTreeMessageTargets(t *testing.T, expectedMsgProp *ExpectedRainTreeMessageProp) {
	addrBook := getAlphabetAddrBook(expectedMsgProp.numNodes)
	network := NewRainTreeNetwork([]byte{expectedMsgProp.orig}, addrBook).(*rainTreeNetwork)
	network.processAddrBookUpdates()

	require.Equal(t, strings.Join(network.addrList, ""), strToAddrList(expectedMsgProp.addrList))

	addrList, addrBookMap, err := network.addrBook.ToListAndMap(network.selfAddr.String())
	require.NoError(t, err)
	require.NotNil(t, addrList)
	require.NotNil(t, addrBookMap)

	i, found := addrList.Find(network.selfAddr.String())
	require.True(t, found)
	require.Equal(t, 0, i)

	for _, target := range expectedMsgProp.targets {
		var addr1, addr2 cryptoPocket.Address
		level := int32(target.level)
		if level == 0 {
			level = getMaxAddrBookLevels(addrBook)
		}
		if level == -1 {
			var ok bool
			addr1, addr2, ok = getLeftAndRight(addrList, addrBookMap)
			assert.True(t, ok)
		}
		if addr1 == nil {
			addr1 = network.getFirstTargetAddr(level)
		}
		require.Equal(t, addr1, cryptoPocket.Address(target.left))

		if addr2 == nil {
			addr2 = network.getSecondTargetAddr(level)
		}
		require.Equal(t, addr2, cryptoPocket.Address(target.right))
	}
}

// Generates an address book with a constant set 27 addresses; ['A', ..., 'Z']
func getAlphabetAddrBook(n int) (addrBook types.AddrBook) {
	addrBook = make([]*types.NetworkPeer, 0)
	for i, ch := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ[" {
		if i >= n {
			return
		}
		addrBook = append(addrBook, &types.NetworkPeer{Address: []byte{byte(ch)}})
	}
	return
}

func strToAddrList(s string) string {
	return hex.EncodeToString([]byte(s))
}