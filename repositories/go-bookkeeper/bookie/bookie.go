package bookie

import "os"

type Bookie interface {
	Start()
}

type Config struct{}

func (c *Config) GetJournalDirs() []os.File {
	return []os.File{}
}

// IsEntryLogPerLedgerEnabled 每个 ledger 单独一个 entry log, 默认 false
// https://bookkeeper.apache.org/docs/reference/config#default-entry-log-settings
func (c *Config) IsEntryLogPerLedgerEnabled() bool {
	return false
}
