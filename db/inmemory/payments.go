package inmemory

import (
	"github.com/bitlum/connector/connectors"
	"sort"
	"sync"
)

// MemoryPaymentsStore is a `PaymentStore` in-memory implementation,
// which is able to periodically clean stored payments which are exceed
// storing time.
type MemoryPaymentsStore struct {
	paymentsMutex sync.RWMutex
	paymentsByID  map[string]*connectors.Payment
}

// Runtime check to ensure that MemoryPaymentsStore implements
// connectors.PaymentsStore interface.
var _ connectors.PaymentsStore = (*MemoryPaymentsStore)(nil)

// NewMemoryPaymentsStore creates new memory payment store with specified
// payments storing time.
func NewMemoryPaymentsStore() *MemoryPaymentsStore {
	return &MemoryPaymentsStore{
		paymentsByID: make(map[string]*connectors.Payment),
	}
}

// PaymentByID return payment by given id.
func (s *MemoryPaymentsStore) PaymentByID(id string) (*connectors.Payment, error) {
	s.paymentsMutex.RLock()
	defer s.paymentsMutex.RUnlock()

	p, exists := s.paymentsByID[id]
	if !exists {
		return nil, connectors.PaymentNotFound
	}
	return p, nil
}

// PaymentByReceipt return payment by given receipt.
func (s *MemoryPaymentsStore) PaymentByReceipt(receipt string) ([]*connectors.Payment, error) {
	s.paymentsMutex.RLock()
	defer s.paymentsMutex.RUnlock()

	var payments []*connectors.Payment
	for _, payment := range s.paymentsByID {
		if payment.Receipt == receipt {
			payments = append(payments, payment)
		}
	}

	sort.Slice(payments, func(i, j int) bool {
		return payments[i].UpdatedAt > payments[j].UpdatedAt
	})

	return payments, nil
}

// SavePayment adds payment to the store.
func (s *MemoryPaymentsStore) SavePayment(p *connectors.Payment) error {
	s.paymentsMutex.Lock()
	defer s.paymentsMutex.Unlock()

	payment := &connectors.Payment{}
	*payment = *p
	s.paymentsByID[p.PaymentID] = payment
	return nil
}

// ListPayments return list of all payments.
func (s *MemoryPaymentsStore) ListPayments(asset connectors.Asset,
	status connectors.PaymentStatus, direction connectors.PaymentDirection,
	media connectors.PaymentMedia, system connectors.PaymentSystem) ([]*connectors.Payment, error) {

	s.paymentsMutex.RLock()
	defer s.paymentsMutex.RUnlock()

	var payments []*connectors.Payment
	for _, payment := range s.paymentsByID {
		if asset != "" && payment.Asset != asset {
			continue
		}

		if status != "" && payment.Status != status {
			continue
		}

		if direction != "" && payment.Direction != direction {
			continue
		}

		if media != "" && payment.Media != media {
			continue
		}

		if system != "" && payment.System != system {
			continue
		}

		payments = append(payments, payment)
	}

	sort.Slice(payments, func(i, j int) bool {
		return payments[i].UpdatedAt > payments[j].UpdatedAt
	})

	return payments, nil
}
