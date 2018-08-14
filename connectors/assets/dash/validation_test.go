package dash

import (
	"testing"
)

func TestValidate(t *testing.T) {

	type args struct {
		asset string
		net   string
		addr  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// DASH mainnet
		{
			name:    "DASH mainnet P2PKH uncompressed",
			args:    args{"DASH", "mainnet", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"},
			wantErr: false,
		},
		{
			name:    "DASH mainnet P2PKH compressed",
			args:    args{"DASH", "mainnet", "Xm1tMRkcJ7u9c95xYeh7vcGj7jrp1EDr54"},
			wantErr: false,
		},
		{
			name:    "DASH mainnet P2PKH hybrid",
			args:    args{"DASH", "mainnet", "XnmpgX9EYz7zFMf5HwLPXbnv9BKqxunaZW"},
			wantErr: false,
		},
		{
			name:    "DASH mainnet P2SH",
			args:    args{"DASH", "mainnet", "7fxExScCeJyW6wmQTu8hxPwWk81dp7LJGf"},
			wantErr: false,
		},
		{
			name:    "DASH mainnet private WIF",
			args:    args{"DASH", "mainnet", "7sL7xs65eEjDatRMr9YusjpFLXoJYjTfjiXqxhXx5ZpfnLdFRAS"},
			wantErr: true,
		},
		{
			name:    "DASH mainnet private WIF",
			args:    args{"DASH", "mainnet", "XK9Pja5RVMbLYZxX1vjrNwyYHCwNT4QhyYK96quc79sQAtJBiPve"},
			wantErr: true,
		},
		{
			name:    "DASH mainnet BTC mainnet address",
			args:    args{"DASH", "mainnet", "1HDNEqzJWdRkcieyD2AHkPJ2wTDW48BpmM"},
			wantErr: true,
		},
		{
			name:    "DASH mainnet BCH mainnet address",
			args:    args{"DASH", "mainnet", "32Y8cHhzt89aZMFKTvDJrfTAdA8VY6rPvp"},
			wantErr: true,
		},
		{
			name:    "DASH mainnet ETH address",
			args:    args{"DASH", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "DASH mainnet LTC mainnet address",
			args:    args{"DASH", "mainnet", "LNfTp5bn61RiCb8AJUEnyJNPqRrqtPAogm"},
			wantErr: true,
		},
		{
			name:    "DASH mainnet regtest address",
			args:    args{"DASH", "mainnet", "yhABgLTC8zqV4ABRTz9xkMnb4A15b8zc57"},
			wantErr: true,
		},
		{
			name:    "DASH mainnet testnet3 address",
			args:    args{"DASH", "mainnet", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"},
			wantErr: true,
		},
		{
			name:    "DASH mainnet random",
			args:    args{"DASH", "mainnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "DASH mainnet empty",
			args:    args{"DASH", "mainnet", ""},
			wantErr: true,
		},

		// DASH regtest
		{
			name:    "DASH regtest P2PKH uncompressed",
			args:    args{"DASH", "regtest", "yhABgLTC8zqV4ABRTz9xkMnb4A15b8zc57"},
			wantErr: false,
		},
		{
			name:    "DASH regtest P2PKH compressed",
			args:    args{"DASH", "regtest", "yWeVNNq3jfZDwt1W7W1Wxdh5Q2MBYQSjwc"},
			wantErr: false,
		},
		{
			name:    "DASH regtest P2PKH hybrid",
			args:    args{"DASH", "regtest", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"},
			wantErr: false,
		},
		{
			name:    "DASH regtest P2SH",
			args:    args{"DASH", "regtest", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"},
			wantErr: false,
		},
		{
			name:    "DASH regtest private WIF",
			args:    args{"DASH", "regtest", "93NB8yoCU8Q9Jbm3aSGtFxoMPXMSJNxKH61cVwcKsW4F46H4p1S"},
			wantErr: true,
		},
		{
			name:    "DASH regtest private WIF",
			args:    args{"DASH", "regtest", "cVSTkDgucjf9egRQNaZ7F3HazQyCfhu9g17hgk2vHuFK2UJvDmFw"},
			wantErr: true,
		},
		{
			name:    "DASH regtest BTC regtest address",
			args:    args{"DASH", "regtest", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"},
			wantErr: true,
		},
		{
			name:    "DASH regtest BCH regtest address",
			args:    args{"DASH", "regtest", "mrQ96nXLKJG4v6XrQRDySwPtBi4QTtRpVU"},
			wantErr: true,
		},
		{
			name:    "DASH regtest ETH address",
			args:    args{"DASH", "regtest", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "DASH regtest LTC regtest address",
			args:    args{"DASH", "regtest", "mixTqvNvpNcuitucquDsXCWxVD6GgihbXQ"},
			wantErr: true,
		},
		{
			name:    "DASH regtest mainnet address",
			args:    args{"DASH", "regtest", "Xm1tMRkcJ7u9c95xYeh7vcGj7jrp1EDr54"},
			wantErr: true,
		},
		{
			name:    "DASH regtest random",
			args:    args{"DASH", "regtest", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "DASH regtest empty",
			args:    args{"DASH", "regtest", ""},
			wantErr: true,
		},
		// false positive
		{
			name:    "DASH regtest testnet3 address",
			args:    args{"DASH", "regtest", "yhABgLTC8zqV4ABRTz9xkMnb4A15b8zc57"},
			wantErr: false,
		},

		// DASH testnet3
		{
			name:    "DASH testnet3 P2PKH uncompressed",
			args:    args{"DASH", "testnet3", "yhABgLTC8zqV4ABRTz9xkMnb4A15b8zc57"},
			wantErr: false,
		},
		{
			name:    "DASH testnet3 P2PKH compressed",
			args:    args{"DASH", "testnet3", "yWeVNNq3jfZDwt1W7W1Wxdh5Q2MBYQSjwc"},
			wantErr: false,
		},
		{
			name:    "DASH testnet3 P2PKH hybrid",
			args:    args{"DASH", "testnet3", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"},
			wantErr: false,
		},
		{
			name:    "DASH testnet3 P2SH",
			args:    args{"DASH", "testnet3", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"},
			wantErr: false,
		},
		{
			name:    "DASH testnet3 private WIF",
			args:    args{"DASH", "testnet3", "93NB8yoCU8Q9Jbm3aSGtFxoMPXMSJNxKH61cVwcKsW4F46H4p1S"},
			wantErr: true,
		},
		{
			name:    "DASH testnet3 private WIF",
			args:    args{"DASH", "testnet3", "cVSTkDgucjf9egRQNaZ7F3HazQyCfhu9g17hgk2vHuFK2UJvDmFw"},
			wantErr: true,
		},
		{
			name:    "DASH testnet3 BTC testnet3 address",
			args:    args{"DASH", "testnet3", "mgG2f5CocNT2bg9hwzuC75mGHZT4tdguXh"},
			wantErr: true,
		},
		{
			name:    "DASH testnet3 BCH testnet3 address",
			args:    args{"DASH", "testnet3", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"},
			wantErr: true,
		},
		{
			name:    "DASH testnet3 ETH address",
			args:    args{"DASH", "testnet3", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "DASH testnet3 LTC testnet4 address",
			args:    args{"DASH", "testnet3", "mixTqvNvpNcuitucquDsXCWxVD6GgihbXQ"},
			wantErr: true,
		},
		{
			name:    "DASH testnet3 mainnet address",
			args:    args{"DASH", "testnet3", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"},
			wantErr: true,
		},
		{
			name:    "DASH testnet3 random",
			args:    args{"DASH", "testnet3", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "DASH testnet3 empty",
			args:    args{"DASH", "testnet3", ""},
			wantErr: true,
		},
		// false positive
		{
			name:    "DASH testnet3 regtest address",
			args:    args{"DASH", "testnet3", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()
			var err error
			if err = ValidateAddress(tt.args.addr, tt.args.net);
				(err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
