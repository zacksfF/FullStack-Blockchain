package worker

import (
	"sync"
	"time"

	"github.com/zacksfF/FullStack-Blockchain/blockchain/database"
	"github.com/zacksfF/FullStack-Blockchain/blockchain/state"
)

// peerUpdateInterval represents the interval of finding new peer nodes
// and updating the blockchain on disk with missing blocks.
const peerUpdateInterval = time.Second * 10

// Worker manages the POW workflows for the blockchain.
type Worker struct {
	state        *state.State
	wg           sync.WaitGroup
	ticker       time.Ticker
	shut         chan struct{}
	startMining  chan bool
	cancelMining chan bool
	txSharing    chan database.BlockTx
	evHandler    state.EventHandler
}

// Run creates a worker, registers the worker with the state package, and
// starts up all the background processes.
func Run(st *state.State, evHandler state.EventHandler) {
	w := Worker{
		state:        st,
		ticker:       *time.NewTicker(peerUpdateInterval),
		shut:         make(chan struct{}),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan bool, 1),
		txSharing:    make(chan database.BlockTx, maxTxShareRequests),
		evHandler:    evHandler,
	}

	// Register this worker with the state package.
	st.Worker = &w

	// Update this node before starting any support G's.
	w.Sync()

	// Select the consensus operation to run.
	consensusOperation := w.powOperations
	if st.Consensus() == state.ConsensusPOA {
		consensusOperation = w.poaOperations
	}

	// Load the set of operations we need to run.
	operations := []func(){
		w.peerOperations,
		w.shareTxOperations,
		consensusOperation,
	}

	// Set waitgroup to match the number of G's we need for the set
	// of operations we have.
	g := len(operations)
	w.wg.Add(g)

	// We don't want to return until we know all the G's are up and running.
	hasStarted := make(chan bool)

	// Start all the operational G's.
	for _, op := range operations {
		go func(op func()) {
			defer w.wg.Done()
			hasStarted <- true
			op()
		}(op)
	}

	// Wait for the G's to report they are running.
	for i := 0; i < g; i++ {
		<-hasStarted
	}
}

// =============================================================================
// These methods implement the state.Worker interface.

// Shutdown terminates the goroutine performing work.
func (w *Worker) Shutdown() {
	w.evHandler("worker: shutdown: started")
	defer w.evHandler("worker: shutdown: completed")

	w.evHandler("worker: shutdown: stop ticker")
	w.ticker.Stop()

	w.evHandler("worker: shutdown: signal cancel mining")
	w.SignalCancelMining()

	w.evHandler("worker: shutdown: terminate goroutines")
	close(w.shut)
	w.wg.Wait()
}

// SignalStartMining starts a mining operation. If there is already a signal
// pending in the channel, just return since a mining operation will start.
func (w *Worker) SignalStartMining() {
	if !w.state.IsMiningAllowed() {
		w.evHandler("state: MinePeerBlock: accepting blocks turned off")
		return
	}

	// Only PoW requires signalling to start mining.
	if w.state.Consensus() != state.ConsensusPOW {
		return
	}

	select {
	case w.startMining <- true:
	default:
	}
	w.evHandler("worker: SignalStartMining: mining signaled")
}

// SignalCancelMining signals the G executing the runMiningOperation function
// to stop immediately.
func (w *Worker) SignalCancelMining() {

	// Only PoW requires signalling to cancel mining.
	if w.state.Consensus() != state.ConsensusPOW {
		return
	}

	select {
	case w.cancelMining <- true:
	default:
	}
	w.evHandler("worker: SignalCancelMining: MINING: CANCEL: signaled")
}

// SignalShareTx signals a share transaction operation. If
// maxTxShareRequests signals exist in the channel, we won't send these.
func (w *Worker) SignalShareTx(blockTx database.BlockTx) {
	select {
	case w.txSharing <- blockTx:
		w.evHandler("worker: SignalShareTx: share Tx signaled")
	default:
		w.evHandler("worker: SignalShareTx: queue full, transactions won't be shared.")
	}
}

// =============================================================================

// isShutdown is used to test if a shutdown has been signaled.
func (w *Worker) isShutdown() bool {
	select {
	case <-w.shut:
		return true
	default:
		return false
	}
}
