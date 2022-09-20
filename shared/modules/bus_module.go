package modules

import (
	"encoding/json"
	"github.com/pokt-network/pocket/shared/debug"
)

// TODO(design): Discuss if this channel should be of pointers to PocketEvents or not. Pointers
// would avoid doing object copying, but might also be less thread safe if another goroutine changes
// it, which could potentially be a feature rather than a bug.
type EventsChannel chan debug.PocketEvent

type Bus interface {
	// Bus Events
	PublishEventToBus(e *debug.PocketEvent)
	GetBusEvent() *debug.PocketEvent
	GetEventBus() EventsChannel

	// Pocket modules
	GetPersistenceModule() PersistenceModule
	GetP2PModule() P2PModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
	GetTelemetryModule() TelemetryModule

	// Configuration
	GetConfig() map[string]json.RawMessage
	GetGenesis() map[string]json.RawMessage
}
