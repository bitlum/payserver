package bitcoin

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
		// BTC mainnet
		{
			name:    "BTC mainnet P2PKH uncompressed",
			args:    args{"BTC", "mainnet", "1PFMrJdc6K61x945CwA7BAYvtVkNoaPcYx"},
			wantErr: false,
		},
		{
			name:    "BTC mainnet P2PKH compressed",
			args:    args{"BTC", "mainnet", "1HDNEqzJWdRkcieyD2AHkPJ2wTDW48BpmM"},
			wantErr: false,
		},
		{
			name:    "BTC mainnet P2PKH hybrid",
			args:    args{"BTC", "mainnet", "1k5N27poM1mpZg6ERvpHAYwRZrMwMt8PX"},
			wantErr: false,
		},
		{
			name:    "BTC mainnet P2WPKH",
			args:    args{"BTC", "mainnet", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tcnxexy2"},
			wantErr: false,
		},
		{
			name:    "BTC mainnet P2SH",
			args:    args{"BTC", "mainnet", "38xPXRp7AZ9XHCnLycRP8rDEeVMG2GYFMg"},
			wantErr: false,
		},
		{
			name:    "BTC mainnet P2WSH",
			args:    args{"BTC", "mainnet", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tc56aenn3yvck7d0x6sc0qgxs65c"},
			wantErr: false,
		},
		{
			name:    "BTC mainnet private WIF uncompressed",
			args:    args{"BTC", "mainnet", "5J879fJS6etub5VKcR8LW6NLhHoAV7a1z4PU1ut5PTYn7xEYJVs"},
			wantErr: true,
		},
		{
			name:    "BTC mainnet private WIF compressed",
			args:    args{"BTC", "mainnet", "KxaN9E3wmxG77Qp19KVP5QhBvq8ks5zjLqxAF2QSN8BG9bxAmyrG"},
			wantErr: true,
		},
		{
			name:    "BTC mainnet ETH mainnet address",
			args:    args{"BTC", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "BTC mainnet LTC mainnet address",
			args:    args{"BTC", "mainnet", "LSN5D48waHCYh5jrLwac1euWydDRtB7M3x"},
			wantErr: true,
		},
		{
			name:    "BTC mainnet DASH mainnet address",
			args:    args{"BTC", "mainnet", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"},
			wantErr: true,
		},
		{
			name:    "BTC mainnet regtest address",
			args:    args{"BTC", "mainnet", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"},
			wantErr: true,
		},
		{
			name:    "BTC mainnet testnet3 address",
			args:    args{"BTC", "mainnet", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"},
			wantErr: true,
		},
		{
			name:    "BTC mainnet simnet address",
			args:    args{"BTC", "mainnet", "rY8j22gBpTbVyp17a3F6eJmGzQRQWLmEcK"},
			wantErr: true,
		},
		{
			name:    "BTC mainnet random",
			args:    args{"BTC", "mainnet", "iGpzgxRNMhDZ1o5sOPNQ6dHmamEXkrlDiA"},
			wantErr: true,
		},
		{
			name:    "BTC mainnet empty",
			args:    args{"BTC", "mainnet", ""},
			wantErr: true,
		},

		// BTC regtest
		{
			name:    "BTC regtest P2PKH uncompressed",
			args:    args{"BTC", "regtest", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"},
			wantErr: false,
		},
		{
			name:    "BTC regtest P2PKH compressed",
			args:    args{"BTC", "regtest", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"},
			wantErr: false,
		},
		{
			name:    "BTC regtest P2PKH hybrid",
			args:    args{"BTC", "regtest", "mgG2f5CocNT2bg9hwzuC75mGHZT4tdguXh"},
			wantErr: false,
		},
		{
			name:    "BTC regtest P2SH",
			args:    args{"BTC", "regtest", "2MzWbbAk8n1esUzQtek3FkoCVrqZRj9kPti"},
			wantErr: false,
		},
		{
			name:    "BTC regtest private WIF uncompressed",
			args:    args{"BTC", "regtest", "91tjjQ7ygsy3Z8zcEm2FNgvJLx9seH7DL1FR6YEajCHptzT74Ye"},
			wantErr: true,
		},
		{
			name:    "BTC regtest private WIF compressed",
			args:    args{"BTC", "regtest", "cNwMc93oD1xNGrHGXjJWSjCFZ4SAXY6RQt6dMSrwsEqGQLx7TAv5"},
			wantErr: true,
		},
		{
			name:    "BTC regtest ETH address",
			args:    args{"BTC", "regtest", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "BTC regtest DASH regtest address",
			args:    args{"BTC", "regtest", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"},
			wantErr: true,
		},
		{
			name:    "BTC regtest mainnet address",
			args:    args{"BTC", "regtest", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tcnxexy2"},
			wantErr: true,
		},
		{
			name:    "BTC regtest simnet address",
			args:    args{"BTC", "regtest", "rY8j22gBpTbVyp17a3F6eJmGzQRQWLmEcK"},
			wantErr: true,
		},
		{
			name:    "BTC regtest random",
			args:    args{"BTC", "regtest", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "BTC regtest empty",
			args:    args{"BTC", "regtest", ""},
			wantErr: true,
		},

		// false positive
		{
			name:    "BTC regtest LTC regtest P2PKH address",
			args:    args{"BTC", "regtest", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"},
			wantErr: false,
		},

		// BTC testnet3
		{
			name:    "BTC testnet3 P2PKH uncompressed",
			args:    args{"BTC", "testnet3", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"},
			wantErr: false,
		},
		{
			name:    "BTC testnet3 P2PKH compressed",
			args:    args{"BTC", "testnet3", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"},
			wantErr: false,
		},
		{
			name:    "BTC testnet3 P2PKH hybrid",
			args:    args{"BTC", "testnet3", "mgG2f5CocNT2bg9hwzuC75mGHZT4tdguXh"},
			wantErr: false,
		},
		{
			name:    "BTC testnet3 P2WPKH",
			args:    args{"BTC", "testnet3", "tb1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tceqz4le"},
			wantErr: false,
		},
		{
			name:    "BTC testnet3 P2SH",
			args:    args{"BTC", "testnet3", "2MzWbbAk8n1esUzQtek3FkoCVrqZRj9kPti"},
			wantErr: false,
		},
		{
			name:    "BTC testnet3 P2WSH",
			args:    args{"BTC", "testnet3", "tb1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tc56aenn3yvck7d0x6sc0qlwx4wh"},
			wantErr: false,
		},
		{
			name:    "BTC testnet3 private WIF uncompressed",
			args:    args{"BTC", "testnet3", "91tjjQ7ygsy3Z8zcEm2FNgvJLx9seH7DL1FR6YEajCHptzT74Ye"},
			wantErr: true,
		},
		{
			name:    "BTC testnet3 private WIF compressed",
			args:    args{"BTC", "testnet3", "cNwMc93oD1xNGrHGXjJWSjCFZ4SAXY6RQt6dMSrwsEqGQLx7TAv5"},
			wantErr: true,
		},
		{
			name:    "BTC testnet3 ETH address",
			args:    args{"BTC", "testnet3", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "BTC testnet3 LTC testnet4 address",
			args:    args{"BTC", "testnet3", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"},
			wantErr: true,
		},
		{
			name:    "BTC testnet3 DASH testnet3 address",
			args:    args{"BTC", "testnet3", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"},
			wantErr: true,
		},
		{
			name:    "BTC testnet3 mainnet address",
			args:    args{"BTC", "testnet3", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tcnxexy2"},
			wantErr: true,
		},
		{
			name:    "BTC testnet3 simnet address",
			args:    args{"BTC", "testnet3", "rY8j22gBpTbVyp17a3F6eJmGzQRQWLmEcK"},
			wantErr: true,
		},
		{
			name:    "BTC testnet3 random",
			args:    args{"BTC", "testnet3", "LZcFOTv5cd0hcMk8vwpK2Mv3kSzRxfjzyT"},
			wantErr: true,
		},
		{
			name:    "BTC testnet3 empty",
			args:    args{"BTC", "testnet3", ""},
			wantErr: true,
		},

		// BTC simnet
		{
			name:    "BTC simnet P2PKH uncompressed",
			args:    args{"BTC", "simnet", "ScgVirScTF4ZeAUiNN8yHx7NWqS21eLuzo"},
			wantErr: true,
		},
		{
			name:    "BTC simnet P2PKH compressed",
			args:    args{"BTC", "simnet", "SSn4xABN9k2yt1HHMZpR5ZSWcNSy7A7bCN"},
			wantErr: true,
		},
		{
			name:    "BTC simnet P2PKH hybrid",
			args:    args{"BTC", "simnet", "Si1sRqSYDbZLMWeDKEwEjpcnCkyjNA1itZ"},
			wantErr: true,
		},
		{
			name:    "BTC simnet P2WPKH ",
			args:    args{"BTC", "simnet", "sb1q3588f3ckfhshjhraeufhe8t82yhmy5auzzmklx"},
			wantErr: true,
		},
		{
			name:    "BTC simnet P2SH",
			args:    args{"BTC", "simnet", "rY8j22gBpTbVyp17a3F6eJmGzQRQWLmEcK"},
			wantErr: true,
		},
		{
			name:    "BTC simnet P2WSH",
			args:    args{"BTC", "simnet", "sb1q3588f3ckfhshjhraeufhe8t82yhmy5aukejx7ntmn6jee9usvyaqus5x4x"},
			wantErr: true,
		},
		{
			name:    "BTC simnet private WIF uncompressed",
			args:    args{"BTC", "simnet", "4N5H6rneNj8wowMxeghGbBHxdF1UEw5dEoR1Bv6NoB5ThBCUgjs"},
			wantErr: true,
		},
		{
			name:    "BTC simnet private WIF compressed",
			args:    args{"BTC", "simnet", "Fr1t4rkTxtz4vAwWdLCjorRrWxq5EVhQwurwqBLjCoUgow4giigq"},
			wantErr: true,
		},
		{
			name:    "BTC simnet ETH address",
			args:    args{"BTC", "simnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "BTC simnet DASH mainnet address",
			args:    args{"BTC", "simnet", "XnmpgX9EYz7zFMf5HwLPXbnv9BKqxunaZW"},
			wantErr: true,
		},
		{
			name:    "BTC simnet mainnet address",
			args:    args{"BTC", "simnet", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tcnxexy2"},
			wantErr: true,
		},
		{
			name:    "BTC simnet testnet3 address",
			args:    args{"BTC", "simnet", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"},
			wantErr: true,
		},
		{
			name:    "BTC simnet random",
			args:    args{"BTC", "simnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "BTC simnet empty",
			args:    args{"BTC", "simnet", ""},
			wantErr: true,
		},
		{
			name:    "BTC simnet LTC simnet address",
			args:    args{"BTC", "simnet", "SdfACz3wafN9FBQNJGi6YwAbMPVqJ8Cc9w"},
			wantErr: true,
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
			if err = ValidateAddress(tt.args.addr, tt.args.net); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
