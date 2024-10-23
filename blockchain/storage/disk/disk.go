package disk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"

	"github.com/zacksfF/FullStack-Blockchain/blockchain/database"
)

// Disk represents the serialization implementation for reading and storing
// blocks in their own separate files on disk. This implements the database.Storage
// interface.
type Disk struct {
	dbPath string
}

// New constructs an Disk value for use.
func New(dbPath string) (*Disk, error) {
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, err
	}

	return &Disk{dbPath: dbPath}, nil
}

// Close in this implementation has nothing to do since a new file is
// written to disk for each now block and then immediately closed.
func (d *Disk) Close() error {
	return nil
}

// Write takes the specified database blocks and stores it on disk in a
// file labeled with the block number.
func (d *Disk) Write(blockData database.BlockData) error {

	// Marshal the block for writing to disk in a more human readable format.
	data, err := json.MarshalIndent(blockData, "", "  ")
	if err != nil {
		return err
	}

	// Create a new file for this block and name it based on the block number.
	f, err := os.OpenFile(d.getPath(blockData.Header.Number), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the new block to disk.
	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

// GetBlock searches the blockchain on disk to locate and return the
// contents of the specified block by number.
func (d *Disk) GetBlock(num uint64) (database.BlockData, error) {

	// Open the block file for the specified number.
	f, err := os.OpenFile(d.getPath(num), os.O_RDONLY, 0600)
	if err != nil {
		return database.BlockData{}, err
	}
	defer f.Close()

	// Decode the contents of the block.
	var blockData database.BlockData
	if err := json.NewDecoder(f).Decode(&blockData); err != nil {
		return database.BlockData{}, err
	}

	// Return the block as a database block.
	return blockData, nil
}

// ForEach returns an iterator to walk through all the blocks
// starting with block number 1.
func (d *Disk) ForEach() database.Iterator {
	return &diskIterator{storage: d}
}

// Reset will clear out the blockchain on disk.
func (d *Disk) Reset() error {
	if err := os.RemoveAll(d.dbPath); err != nil {
		return err
	}

	return os.MkdirAll(d.dbPath, 0755)
}

// getPath forms the path to the specified block.
func (d *Disk) getPath(blockNum uint64) string {
	name := strconv.FormatUint(blockNum, 10)
	return path.Join(d.dbPath, fmt.Sprintf("%s.json", name))
}

// =============================================================================

// diskIterator represents the iteration implementation for walking
// through and reading blocks on disk. This implements the database
// Iterator interface.
type diskIterator struct {
	storage *Disk  // Access to the storage API.
	current uint64 // Currenet block number being iterated over.
	eoc     bool   // Represents the iterator is at the end of the chain.
}

// Next retrieves the next block from disk.
func (di *diskIterator) Next() (database.BlockData, error) {
	if di.eoc {
		return database.BlockData{}, errors.New("end of chain")
	}

	di.current++
	blockData, err := di.storage.GetBlock(di.current)
	if errors.Is(err, fs.ErrNotExist) {
		di.eoc = true
	}

	return blockData, err
}

// Done returns the end of chain value.
func (di *diskIterator) Done() bool {
	return di.eoc
}
