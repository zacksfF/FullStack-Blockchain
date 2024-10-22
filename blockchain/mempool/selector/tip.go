package selector

import (
	"sort"

	"github.com/zacksfF/FullStack-Blockchain/blockchain/database"
)

// CORE NOTE: On Ethereum a transaction will stay in the mempool and not be selected
// unless the transaction holds the next expected nonce. Transactions can get stuck
// in the mempool because of this. This is very complicated for us to implement for
// now. So we will check the nonce for each transaction when the block is mined.
// If the nonce is not expected, it will fail but the user continues to pay fees.

// tipSelect returns transactions with the best tip while respecting the nonce
// for each account/transaction.
var tipSelect = func(m map[database.AccountID][]database.BlockTx, howMany int) []database.BlockTx {

	/*
		Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
			  {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		Pavl: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
			  {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		Edua: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
			  {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
	*/

	// Sort the transactions per account by nonce.
	for key := range m {
		if len(m[key]) > 1 {
			sort.Sort(byNonce(m[key]))
		}
	}

	/*
		Bill: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		      {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
		Pavl: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		      {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
		Edua: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		      {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
	*/

	// Pick the first transaction in the slice for each account. Each iteration
	// represents a new row of selections. Keep doing that until all the
	// transactions have been selected.
	var rows [][]database.BlockTx
	for {
		var row []database.BlockTx
		for key := range m {
			if len(m[key]) > 0 {
				row = append(row, m[key][0])
				m[key] = m[key][1:]
			}
		}
		if row == nil {
			break
		}
		rows = append(rows, row)
	}

	/*
		0: Bill: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		0: Pavl: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		0: Edua: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		1: Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
		1: Pavl: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
		1: Edua: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
	*/

	// Sort each row by tip unless we will take all transactions from that row
	// anyway. Then try to select the number of requested transactions. Keep
	// pulling transactions from each row until the amount of fulfilled or
	// there are no more transactions.
	final := []database.BlockTx{}
	for _, row := range rows {
		need := howMany - len(final)
		if len(row) > need {
			sort.Sort(byTip(row))
			final = append(final, row[:need]...)
			break
		}
		final = append(final, row...)
	}

	/*
		0: Bill: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		1: Pavl: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		2: Edua: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		3: Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
	*/

	return final
}
