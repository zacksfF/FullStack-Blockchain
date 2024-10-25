package private

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/zacksfF/FullStack-Blockchain/blockchain/database"
	"github.com/zacksfF/FullStack-Blockchain/blockchain/peer"
	"github.com/zacksfF/FullStack-Blockchain/blockchain/state"
	"github.com/zacksfF/FullStack-Blockchain/core/web/errs"
	"github.com/zacksfF/FullStack-Blockchain/nameservices"
	"github.com/zacksfF/FullStack-Blockchain/web2"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservices.NameService
}

// SubmitNodeTransaction adds new node transactions to the mempool.
func (h Handlers) SubmitNodeTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web2.GetValues(ctx)
	if err != nil {
		return web2.NewShutdownError("web value missing from context")
	}

	// Decode the JSON in the post call into a block transaction.
	var tx database.BlockTx
	if err := web2.Decode(r, &tx); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	// Ask the state package to add this transaction to the mempool and perform
	// any other business logic.
	h.Log.Infow("add tran", "traceid", v.TraceID, "sig:nonce", tx, "fron", tx.FromID, "to", tx.ToID, "value", tx.Value, "tip", tx.Tip)
	if err := h.State.UpsertNodeTransaction(tx); err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "transactions added to mempool",
	}

	return web2.Respond(ctx, w, resp, http.StatusOK)
}

// ProposeBlock takes a block received from a peer, validates it and
// if that passes, adds the block to the local blockchain.
func (h Handlers) ProposeBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	// Decode the JSON in the post call into a file system block.
	var blockData database.BlockData
	if err := web2.Decode(r, &blockData); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	// Convert the block data into a block. This action will create a merkle
	// tree for the set of transactions required for blockchain operations.
	block, err := database.ToBlock(blockData)
	if err != nil {
		return fmt.Errorf("unable to decode block: %w", err)
	}

	// Ask the state package to validate the proposed block. If the block
	// passes validation, it will be added to the blockchain database.
	if err := h.State.ProcessProposedBlock(block); err != nil {
		if errors.Is(err, database.ErrChainForked) {
			h.State.Reorganize()
		}

		return errs.NewTrusted(errors.New("block not accepted"), http.StatusNotAcceptable)
	}

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "accepted",
	}

	return web2.Respond(ctx, w, resp, http.StatusOK)
}

// SubmitPeer is called by a node so they can be added to the known peer list.
func (h Handlers) SubmitPeer(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web2.GetValues(ctx)
	if err != nil {
		return web2.NewShutdownError("web value missing from context")
	}

	var peer peer.Peer
	if err := web2.Decode(r, &peer); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	if !h.State.AddKnownPeer(peer) {
		h.Log.Infow("adding peer", "traceid", v.TraceID, "host", peer.Host)
	}

	return web2.Respond(ctx, w, nil, http.StatusOK)
}

// Status returns the current status of the node.
func (h Handlers) Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	latestBlock := h.State.LatestBlock()

	status := peer.PeerStatus{
		LatestBlockHash:   latestBlock.Hash(),
		LatestBlockNumber: latestBlock.Header.Number,
		KnownPeers:        h.State.KnownExternalPeers(),
	}

	return web2.Respond(ctx, w, status, http.StatusOK)
}

// BlocksByNumber returns all the blocks based on the specified to/from values.
func (h Handlers) BlocksByNumber(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	fromStr := web2.Param(r, "from")
	if fromStr == "latest" || fromStr == "" {
		fromStr = fmt.Sprintf("%d", state.QueryLastest)
	}

	toStr := web2.Param(r, "to")
	if toStr == "latest" || toStr == "" {
		toStr = fmt.Sprintf("%d", state.QueryLastest)
	}

	from, err := strconv.ParseUint(fromStr, 10, 64)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}
	to, err := strconv.ParseUint(toStr, 10, 64)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	if from > to {
		return errs.NewTrusted(errors.New("from greater than to"), http.StatusBadRequest)
	}

	blocks := h.State.QueryBlocksByNumber(from, to)
	if len(blocks) == 0 {
		return web2.Respond(ctx, w, nil, http.StatusNoContent)
	}

	blockData := make([]database.BlockData, len(blocks))
	for i, block := range blocks {
		blockData[i] = database.NewBlockData(block)
	}

	return web2.Respond(ctx, w, blockData, http.StatusOK)
}

// Mempool returns the set of uncommitted transactions.
func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	txs := h.State.Mempool()
	return web2.Respond(ctx, w, txs, http.StatusOK)
}
