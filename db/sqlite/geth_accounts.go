package sqlite

import (
	"github.com/bitlum/connector/connectors/daemons/geth"
	"github.com/jinzhu/gorm"
	"time"
)

type EthereumState struct {
	DefaultAddressNonce int
}

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
	s.db.globalMutex.Lock()
	defer s.db.globalMutex.Unlock()

	address := &EthereumAddress{}
	err := s.db.Where("address = ?", addressStr).Find(address).Error
	if gorm.IsRecordNotFoundError(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return address.Account, nil
}

// GetAddressesByAccount returns addressed belonging to the given account.
//
// NOTE: Part of the geth.AccountsStorage interface.
func (s *GethAccountsStorage) GetAddressesByAccount(accountStr string) (
	[]string, error) {

	s.db.globalMutex.Lock()
	defer s.db.globalMutex.Unlock()

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
	s.db.globalMutex.Lock()
	defer s.db.globalMutex.Unlock()

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
	s.db.globalMutex.Lock()
	defer s.db.globalMutex.Unlock()

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
	s.db.globalMutex.Lock()
	defer s.db.globalMutex.Unlock()

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

// PutDefaultAddressNonce puts returns default address transaction nonce.
// This method is needed because if we send transaction too frequently
// ethereum transaction counter couldn't keep up and transaction fails,
// because of replacement error.
//
// NOTE: Part of the geth.AccountsStorage interface.
func (s *GethAccountsStorage) PutDefaultAddressNonce(nonce int) error {
	s.db.globalMutex.Lock()
	defer s.db.globalMutex.Unlock()

	return s.db.Save(&EthereumState{
		DefaultAddressNonce: nonce,
	}).Error
}

// DefaultAddressNonce returns default address transaction nonce.
// This method is needed because if we send transaction too frequently
// ethereum transaction counter couldn't keep up and transaction fails,
// because of replacement error.
func (s *GethAccountsStorage) DefaultAddressNonce() (int, error) {
	s.db.globalMutex.Lock()
	defer s.db.globalMutex.Unlock()

	state := &EthereumState{}
	if err := s.db.Find(state).Error; err != nil {
		return 0, err
	}

	return state.DefaultAddressNonce, nil
}
