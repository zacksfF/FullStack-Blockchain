package genesis

import "time"

//Package genesis maintains access to the genesis file

//Genesis reprsent the genesis file.
type Genesis struct{
	Date time.Time `json:"date"`
	ChainID uint16 
}