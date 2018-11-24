package sqlite

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
	"github.com/bitlum/connector/connectors"
)

func migrate(gdb *gorm.DB) error {
	migrations := gormigrate.New(gdb, gormigrate.DefaultOptions,
		[]*gormigrate.Migration{
			{
				ID: "add_payment_system_type",
				Migrate: func(tx *gorm.DB) error {
					if err := tx.AutoMigrate(&Payment{}).Error; err != nil {
						return err
					}

					store := PaymentsStore{DB: &DB{DB: tx}}

					// Previous internal payment direction now should be
					// moved in system field. All previous internal
					// payment where incoming.
					payments, err := store.ListPayments("", "",
						"internal", "", "")
					if err != nil {
						return err
					}

					for _, payment := range payments {
						payment.System = connectors.Internal
						payment.Direction = connectors.Incoming
						if err := tx.Save(payment).Error; err != nil {
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
						if err := tx.Save(payment).Error; err != nil {
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
						if err := tx.Save(payment).Error; err != nil {
							return err
						}
					}

					return nil
				},
			},
		})

	return migrations.Migrate()
}
