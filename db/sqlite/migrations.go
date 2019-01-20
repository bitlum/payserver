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
		&BitcoinSimpleState{},
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
		store := PaymentsStore{db: &DB{DB: tx}}

		// Previous internal payment direction now should be
		// moved in system field. All previous internal
		// payment where incoming.
		payments, err := store.ListPayments("", "",
			"", "", "")
		if err != nil {
			return err
		}

		for _, payment := range payments {
			// Internal transaction were tracked unproperly previously.
			if payment.Direction == "Internal" {
				err := tx.Delete(&Payment{}, "payment_id = ?",
					payment.PaymentID).Error
				if err != nil {
					return err
				}

				continue
			}

			// All other transactions were external
			oldID := payment.PaymentID
			err := tx.Delete(&Payment{}, "payment_id = ?", oldID).Error
			if err != nil {
				return err
			}

			payment.System = connectors.External
			payment.PaymentID, err = payment.GenPaymentID()
			if err != nil {
				return err
			}

			log.Infof("Payment migration (%v) => (%v)", oldID, payment.PaymentID)
			if err := store.SavePayment(payment); err != nil {
				return err
			}
		}

		return nil
	},
}
