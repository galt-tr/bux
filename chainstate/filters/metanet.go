package filters

import (
	"strings"

	"github.com/libsv/go-bt"
	"github.com/mrz1836/go-whatsonchain"
)

// MetanetScriptTemplate script template for metanet transaction
const MetanetScriptTemplate = "14c91e5cc393bb9d6da3040a7c72b4b569b237e450"

// Metanet filter processor
func Metanet(tx *whatsonchain.TxInfo) (*bt.Tx, error) {
	// Loop through all of the outputs and check for pubkeyhash output
	for _, out := range tx.Vout {
		// if any output contains a pubkeyhash output, include this tx in the filter
		if strings.HasPrefix(out.ScriptPubKey.Hex, MetanetScriptTemplate) {
			return bt.NewTxFromString(tx.Hex)
		}
	}
	return nil, nil
}
