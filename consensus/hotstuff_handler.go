package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
)

// DISCUSS: Should these functions return an error?
type HotstuffMessageHandler interface {
	HandleNewRoundMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandlePrepareMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandlePrecommitMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandleCommitMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandleDecideMessage(*ConsensusModule, *typesCons.HotstuffMessage)
}

func (m *ConsensusModule) handleHotstuffMessage(msg *typesCons.HotstuffMessage) error {
	m.nodeLog(typesCons.DebugHandlingHotstuffMessage(msg))

	step := msg.GetStep()

	// Pacemaker - Liveness & safety checks
	if err := m.paceMaker.ValidateMessage(msg); err != nil {
		// If a replica is not a leader for this round, but has already determined a leader,
		// and continues to receive NewRound messages, we avoid logging the "message discard"
		// because it creates unnecessary spam.
		if !(m.LeaderId != nil && !m.isLeader() && step == NewRound) {
			m.nodeLog(typesCons.WarnDiscardHotstuffMessage(msg, err.Error()))
		}
		return err
	}

	// Leader Election - Need to execute leader election if there is no leader and we are in a new round.
	if m.Step == NewRound && m.LeaderId == nil {
		if err := m.electNextLeader(msg); err != nil {
			return err
		}
	}

	// Hotstuff - Handle message
	if m.isReplica() {
		replicaHandlers[step](m, msg)
	}
	// Note that the leader also acts as a replica, but this logic is implemented in the underlying code.
	leaderHandlers[step](m, msg)

	return nil
}