package sqlite

import (
	"github.com/bitlum/connector/connectors"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func (db *DB) Migrate() error {
	if err := db.DB.AutoMigrate(
		&EthereumState{},
		&ConnectorState{},
		&EthereumAddress{},
		&Payment{},
	).Error; err != nil {
		return err
	}

	return migrate(db.DB, allMigrations)
}

func migrate(gdb *gorm.DB, migrations []*gormigrate.Migration) error {
	return gormigrate.New(gdb, gormigrate.DefaultOptions, migrations).Migrate()
}

var allMigrations = []*gormigrate.Migration{
	addPaymentSystemType,
}

var addPaymentSystemType = &gormigrate.Migration{
	ID: "add_payment_system_type",
	Migrate: func(tx *gorm.DB) error {
		store := PaymentsStore{DB: &DB{DB: tx}}

		// Previous internal payment direction now should be
		// moved in system field. All previous internal
		// payment where incoming.
		payments, err := store.ListPayments("", "",
			"Internal", "", "")
		if err != nil {
			return err
		}

		for _, payment := range payments {
			payment.System = connectors.Internal
			payment.Direction = connectors.Incoming
			if err := store.SavePayment(payment); err != nil {
				return err
			}
		}

		// Previous incoming and outgoing payments were external.
		payments, err = store.ListPayments("", "",
			connectors.Outgoing, "", "")
		if err != nil {
			return err
		}

		for _, payment := range payments {
			payment.System = connectors.External
			err := tx.Delete(&Payment{}, "payment_id = ?",
				payment.PaymentID).Error
			if err != nil {
				return err
			}

			payment.PaymentID, err = payment.GenPaymentID()
			if err != nil {
				return err
			}

			if err := store.SavePayment(payment); err != nil {
				return err
			}
		}

		payments, err = store.ListPayments("", "",
			connectors.Incoming, "", "")
		if err != nil {
			return err
		}

		for _, payment := range payments {
			payment.System = connectors.Internal
			err := tx.Delete(&Payment{}, "payment_id = ?",
				payment.PaymentID).Error
			if err != nil {
				return err
			}

			payment.PaymentID, err = payment.GenPaymentID()
			if err != nil {
				return err
			}

			if err := store.SavePayment(payment); err != nil {
				return err
			}
		}

		return nil
	},
}
