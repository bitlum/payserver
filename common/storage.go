package common

import (
	"errors"
	"sync"
	"time"

	"github.com/bitlum/connector/connectors"
)

// PaymentStore is used to store payments for external services for give
// them ability to check whether payment is received.
type PaymentsStore interface {
	// Payment returns payment by id.
	Payment(id string) (*connectors.Payment, error)

	// AddPayment add payment to the store.
	AddPayment(p *connectors.Payment) error
}

var PaymentNotFound = errors.New("payment not found")

// MemoryPaymentsStore is a `PaymentStore` in-memory implementation,
// which is able to periodically clean stored payments which are exceed
// storing time.
type MemoryPaymentsStore struct {
	paymentsMutex    sync.RWMutex
	payments         map[string]paymentWithDatetime
	storingTime      time.Duration
	cleanerStopper   chan struct{}
	cleanerWaitGroup sync.WaitGroup
}

// NewMemoryPaymentsStore creates new memory payment store with specified
// payments storing time.
func NewMemoryPaymentsStore(storingTime time.Duration) *MemoryPaymentsStore {
	return &MemoryPaymentsStore{
		payments:    map[string]paymentWithDatetime{},
		storingTime: storingTime,
	}
}

type paymentWithDatetime struct {
	*connectors.Payment
	datetime time.Time
}

// Payment return payment with specified id.
func (s *MemoryPaymentsStore) Payment(id string) (*connectors.Payment, error) {
	s.paymentsMutex.RLock()
	defer s.paymentsMutex.RUnlock()
	p, exists := s.payments[id]
	if !exists {
		return nil, PaymentNotFound
	}
	return p.Payment, nil
}

// AddPayment adds payment to the store.
func (s *MemoryPaymentsStore) AddPayment(p *connectors.Payment) error {
	s.paymentsMutex.Lock()
	defer s.paymentsMutex.Unlock()
	s.payments[p.ID] = paymentWithDatetime{
		Payment:  p,
		datetime: time.Now(),
	}
	return nil
}

// StartCleaner starts cleaner which periodically removes outdated
// payments.
func (s *MemoryPaymentsStore) StartCleaner() error {
	if s.cleanerStopper != nil {
		return errors.New("already started")
	}
	s.cleanerWaitGroup.Add(1)
	go func() {
		defer s.cleanerWaitGroup.Done()
		if s.cleanerStopper == nil {
			return
		}
		for {
			select {
			case <-time.After(time.Minute):
				s.cleanPayments()
			case <-s.cleanerStopper:
				return
			}
		}
	}()
	return nil
}

// StopCleaner stops periodical outdated payments cleaning.
func (s *MemoryPaymentsStore) StopCleaner() error {
	if s.cleanerStopper == nil {
		return errors.New("not started yet or already stopped")
	}
	close(s.cleanerStopper)
	s.cleanerWaitGroup.Wait()
	s.cleanerStopper = nil
	return nil
}

// cleanPayments directly cleans outdated payments. Intended to be called
// in cleaner goroutine.
func (s *MemoryPaymentsStore) cleanPayments() {
	s.paymentsMutex.Lock()
	defer s.paymentsMutex.Unlock()
	for id, p := range s.payments {
		if p.datetime.Add(s.storingTime).Before(time.Now()) {
			delete(s.payments, id)
		}
	}
}
