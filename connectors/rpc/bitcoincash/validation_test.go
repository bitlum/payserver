package bitcoincash

import "testing"

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
		// BCH mainnet
		{
			name:    "BCH mainnet P2PKH uncompressed",
			args:    args{"BCH", "mainnet", "1K6aphb1obCKoLSfL7KZyvBS6hogcUzZNy"},
			wantErr: false,
		},
		{
			name:    "BCH mainnet P2PKH compressed",
			args:    args{"BCH", "mainnet", "1BtBojSMWGpp8z4EgrFbd2BZKiThXRYX1e"},
			wantErr: false,
		},
		{
			name:    "BCH mainnet P2PKH hybrid",
			args:    args{"BCH", "mainnet", "1GmibqMsdE5jLFXGaoEX5gG4QAf8XX4uMZ"},
			wantErr: false,
		},
		{
			name:    "BCH mainnet P2PKH CashAddr uncompressed",
			args:    args{"BCH", "mainnet", "bitcoincash:qrrgpy7nffggd9g0fen82lrhtemauurtnuq46g5jl7"},
			wantErr: false,
		},
		{
			name:    "BCH mainnet P2PKH CashAddr compressed",
			args:    args{"BCH", "mainnet", "bitcoincash:qpm47l0kukuzjnk2vsp70256s9pd99qs5u2e7gd5f7"},
			wantErr: false,
		},
		{
			name:    "BCH mainnet P2PKH CashAddr hybrid",
			args:    args{"BCH", "mainnet", "bitcoincash:qzk0a6jahmytcdqeh56uj6pe036fkcf9rq45kwptuk"},
			wantErr: false,
		},
		{
			name:    "BCH mainnet P2SH",
			args:    args{"BCH", "mainnet", "32Y8cHhzt89aZMFKTvDJrfTAdA8VY6rPvp"},
			wantErr: false,
		},
		{
			name:    "BCH mainnet private WIF uncompressed",
			args:    args{"BCH", "mainnet", "5KTkH5jkTaDgMfww9R9uUvXtsc8N1rqntvAdVkhXRhoQPtudYuu"},
			wantErr: true,
		},
		{
			name:    "BCH mainnet private WIF compressed",
			args:    args{"BCH", "mainnet", "L4V41yugpHcxiEuTz9cJeFHMxN3VknzMa9fW5ussJ35pG4BgjFSf"},
			wantErr: true,
		},
		{
			name:    "BCH mainnet ETH address",
			args:    args{"BCH", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "BCH mainnet LTC mainnet address",
			args:    args{"BCH", "mainnet", "LSN5D48waHCYh5jrLwac1euWydDRtB7M3x"},
			wantErr: true,
		},
		{
			name:    "BCH mainnet DASH mainnet address",
			args:    args{"BCH", "mainnet", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"},
			wantErr: true,
		},
		{
			name:    "BCH mainnet regtest address",
			args:    args{"BCH", "mainnet", "mycY7kfzccdaaSvH3gHwoqPkxhQPXVzSwz"},
			wantErr: true,
		},
		{
			name:    "BCH mainnet testnet3 address",
			args:    args{"BCH", "mainnet", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"},
			wantErr: true,
		},
		{
			name:    "BCH mainnet random",
			args:    args{"BCH", "mainnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "BCH mainnet empty",
			args:    args{"BCH", "mainnet", ""},
			wantErr: true,
		},

		// BCH testnet3
		{
			name:    "BCH testnet3 P2PKH uncompressed",
			args:    args{"BCH", "testnet3", "mycY7kfzccdaaSvH3gHwoqPkxhQPXVzSwz"},
			wantErr: false,
		},
		{
			name:    "BCH testnet3 P2PKH compressed",
			args:    args{"BCH", "testnet3", "mrQ96nXLKJG4v6XrQRDySwPtBi4QTtRpVU"},
			wantErr: false,
		},
		{
			name:    "BCH testnet3 P2PKH hybrid",
			args:    args{"BCH", "testnet3", "mwHfttSrSFWz7MztJNCtubUPGAFqRPHUKi"},
			wantErr: false,
		},
		{
			name:    "BCH testnet3 P2PKH CashAddr uncompressed",
			args:    args{"BCH", "testnet3", "bchtest:qrrgpy7nffggd9g0fen82lrhtemauurtnuy870k9cz"},
			wantErr: false,
		},
		{
			name:    "BCH testnet3 P2PKH CashAddr compressed",
			args:    args{"BCH", "testnet3", "bchtest:qpm47l0kukuzjnk2vsp70256s9pd99qs5uwt600rwz"},
			wantErr: false,
		},
		{
			name:    "BCH testnet3 P2PKH CashAddr hybrid",
			args:    args{"BCH", "testnet3", "bchtest:qzk0a6jahmytcdqeh56uj6pe036fkcf9rq3xjfrum2"},
			wantErr: false,
		},
		{
			name:    "BCH testnet3 P2SH",
			args:    args{"BCH", "testnet3", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"},
			wantErr: false,
		},
		{
			name:    "BCH testnet3 private WIF uncompressed",
			args:    args{"BCH", "testnet3", "93ENrpZJ3oHpKjTDmm3pMX5rXGV5B2NzEs2aaP42mSYTAwLB77a"},
			wantErr: true,
		},
		{
			name:    "BCH testnet3 private WIF compressed",
			args:    args{"BCH", "testnet3", "cUr3UtuYFMKDsgNjNZRS1ZnRabLuRF63eBoyCLLNo9jpWoH6Mxw2"},
			wantErr: true,
		},
		{
			name:    "BCH testnet3 ETH address",
			args:    args{"BCH", "testnet3", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "BCH testnet3 LTC testnet4 address",
			args:    args{"BCH", "testnet3", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"},
			wantErr: true,
		},
		{
			name:    "BCH testnet3 DASH testnet3 address",
			args:    args{"BCH", "testnet3", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"},
			wantErr: true,
		},
		{
			name:    "BCH testnet3 random",
			args:    args{"BCH", "testnet3", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "BCH testnet3 empty",
			args:    args{"BCH", "testnet3", ""},
			wantErr: true,
		},

		// BCH regtest
		{
			name:    "BCH regtest P2PKH uncompressed",
			args:    args{"BCH", "regtest", "mycY7kfzccdaaSvH3gHwoqPkxhQPXVzSwz"},
			wantErr: false,
		},
		{
			name:    "BCH regtest P2PKH compressed",
			args:    args{"BCH", "regtest", "mrQ96nXLKJG4v6XrQRDySwPtBi4QTtRpVU"},
			wantErr: false,
		},
		{
			name:    "BCH regtest P2PKH hybrid",
			args:    args{"BCH", "regtest", "mwHfttSrSFWz7MztJNCtubUPGAFqRPHUKi"},
			wantErr: false,
		},
		{
			name:    "BCH regtest P2PKH CashAddr uncompressed",
			args:    args{"BCH", "regtest", "bchtest:qrrgpy7nffggd9g0fen82lrhtemauurtnuy870k9cz"},
			wantErr: false,
		},
		{
			name:    "BCH regtest P2PKH CashAddr compressed",
			args:    args{"BCH", "regtest", "bchtest:qpm47l0kukuzjnk2vsp70256s9pd99qs5uwt600rwz"},
			wantErr: false,
		},
		{
			name:    "BCH regtest P2PKH CashAddr hybrid",
			args:    args{"BCH", "regtest", "bchtest:qzk0a6jahmytcdqeh56uj6pe036fkcf9rq3xjfrum2"},
			wantErr: false,
		},
		{
			name:    "BCH regtest P2SH",
			args:    args{"BCH", "regtest", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"},
			wantErr: false,
		},
		{
			name:    "BCH regtest private WIF uncompressed",
			args:    args{"BCH", "regtest", "93ENrpZJ3oHpKjTDmm3pMX5rXGV5B2NzEs2aaP42mSYTAwLB77a"},
			wantErr: true,
		},
		{
			name:    "BCH regtest private WIF compressed",
			args:    args{"BCH", "regtest", "cUr3UtuYFMKDsgNjNZRS1ZnRabLuRF63eBoyCLLNo9jpWoH6Mxw2"},
			wantErr: true,
		},
		{
			name:    "BCH regtest ETH address",
			args:    args{"BCH", "regtest", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "BCH regtest DASH regtest address",
			args:    args{"BCH", "regtest", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"},
			wantErr: true,
		},
		{
			name:    "BCH regtest random",
			args:    args{"BCH", "regtest", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "BCH regtest empty",
			args:    args{"BCH", "regtest", ""},
			wantErr: true,
		},
		// false positive
		{
			name:    "BCH regtest LTC mainnet P2PKH address",
			args:    args{"BCH", "regtest", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"},
			wantErr: false,
		},
		{
			name:    "BCH mainnet BCH mainnet P2PKH address",
			args:    args{"BCH", "mainnet", "qrpu9ylkzk5jxq3nl43d0jndmyc4el4qxgunewsufk"},
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
			if _, err = DecodeAddress(tt.args.addr,
				tt.args.net); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
