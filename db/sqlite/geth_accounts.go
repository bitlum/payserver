package sqlite

import (
	"github.com/bitlum/connector/connectors/daemons/geth"
	"time"
	"github.com/jinzhu/gorm"
	"github.com/davecgh/go-spew/spew"
)

type EthereumAddress struct {
	CreatedAt time.Time

	Address string `gorm:"primary_key"`
	Account string
}

// GethAccountsStorage is used to keep track connections between addresses and
// accounts, because of the reason of Ethereum client not having this mapping
// internally.
type GethAccountsStorage struct {
	db *DB
}

func NewGethAccountsStorage(db *DB) *GethAccountsStorage {
	return &GethAccountsStorage{
		db: db,
	}
}

// Runtime check to ensure that GethAccountsStorage implements
// geth.AccountsStorage interface.
var _ geth.AccountsStorage = (*GethAccountsStorage)(nil)

// GetAccountByAddress returns account by given address.
//
// NOTE: Part of the geth.AccountsStorage interface.
func (s *GethAccountsStorage) GetAccountByAddress(addressStr string) (string, error) {
	address := &EthereumAddress{Address: addressStr}
	err := s.db.Find(address).Error
	if gorm.IsRecordNotFoundError(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	spew.Dump(address)

	return address.Account, nil
}

// GetAddressesByAccount returns addressed belonging to the given account.
//
// NOTE: Part of the geth.AccountsStorage interface.
func (s *GethAccountsStorage) GetAddressesByAccount(accountStr string) (
	[]string, error) {

	var ethereumAddresses []*EthereumAddress
	err := s.db.Find(&ethereumAddresses, "account = ?", accountStr).Error
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, address := range ethereumAddresses {
		addresses = append(addresses, address.Address)
	}

	return addresses, nil
}

// GetLastAccountAddress returns last address which were assigned to
// account.
//
// NOTE: Part of the geth.AccountsStorage interface.
func (s *GethAccountsStorage) GetLastAccountAddress(accountStr string) (string,
	error) {
	ethereumAddresses := make([]*EthereumAddress, 0)
	err := s.db.Where("account = ?", accountStr).Find(&ethereumAddresses).Error
	if err != nil {
		return "", err
	}

	if len(ethereumAddresses) > 0 {
		return ethereumAddresses[len(ethereumAddresses)-1].Address, nil
	} else {
		return "", nil
	}
}

// AddAddressToAccount assigns new address to account.
//
// NOTE: Part of the geth.AccountsStorage interface.
func (s *GethAccountsStorage) AddAddressToAccount(addressStr,
accountStr string) error {
	address := &EthereumAddress{
		Account: accountStr,
		Address: addressStr,
	}

	return s.db.Save(address).Error
}

// AllAccounts returns all created accounts.
//
// NOTE: Part of the geth.AccountsStorage interface.
func (s *GethAccountsStorage) AllAddresses() ([]string, error) {
	var ethereumAddresses []*EthereumAddress
	err := s.db.Find(&ethereumAddresses).Error
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, address := range ethereumAddresses {
		addresses = append(addresses, address.Address)
	}

	return addresses, nil
}
