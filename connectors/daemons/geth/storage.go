package geth

// AccountsStorage is used to keep track connections between addresses and
// accounts, because of the reason of Ethereum client not having this mapping
// internally.
//
// NOTE: This storage has to be persistent.
type AccountsStorage interface {
	// GetAccountByAddress returns account by given address.
	GetAccountByAddress(address string) (string, error)

	// GetAddressesByAccount returns addressed belonging to the given account.
	GetAddressesByAccount(account string) ([]string, error)

	// GetLastAccountAddress returns last address which were assigned to
	// account.
	GetLastAccountAddress(account string) (string, error)

	// AddAddressToAccount assigns new address to account.
	AddAddressToAccount(address, account string) error

	// AllAddresses returns all created addresses.
	AllAddresses() ([]string, error)
}
