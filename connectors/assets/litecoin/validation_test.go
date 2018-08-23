package litecoin

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
		// LTC mainnet
		{
			name:    "LTC mainnet P2PKH uncompressed",
			args:    args{"LTC", "mainnet", "LQasau6s59W354CqQk7Hmks5E7XagUMwxj"},
			wantErr: false,
		},
		{
			name:    "LTC mainnet P2PKH compressed",
			args:    args{"LTC", "mainnet", "LXQBaiuzH5UqN1P2MSaNxJ4iE1EBUNn19c"},
			wantErr: false,
		},
		{
			name:    "LTC mainnet P2PKH hybrid",
			args:    args{"LTC", "mainnet", "LhB4WoN5c9Btx9CRd2eqEsCBbYW6met5mb"},
			wantErr: false,
		},
		{
			name:    "LTC mainnet P2WPKH ",
			args:    args{"LTC", "mainnet", "ltc1qupndfjxttgfdtq3k4wzuvyegcdz8uun0t09j0n"},
			wantErr: false,
		},
		{
			name:    "LTC mainnet P2SH legacy",
			args:    args{"LTC", "mainnet", "35fs1NJAvtMvL2EsAzFPwtdyEmQk2LTBHs"},
			wantErr: false,
		},
		{
			name:    "LTC mainnet P2SH",
			args:    args{"LTC", "mainnet", "ZG26rWenBR7dfXhN1TEJKJ3ySbjVPismuT"},
			wantErr: false,
		},
		{
			name:    "LTC mainnet P2WSH",
			args:    args{"LTC", "mainnet", "ltc1qupndfjxttgfdtq3k4wzuvyegcdz8uun0lwa47m8k3e3qvcw4wuusn8p8rk"},
			wantErr: false,
		},
		{
			name:    "LTC mainnet private WIF uncompressed",
			args:    args{"LTC", "mainnet", "6vDkyaqgbDCf67Em5jsJzFcZvPucRf2Z7pHvj9UhehmWTqFigeL"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet private WIF compressed",
			args:    args{"LTC", "mainnet", "T7uK6kMsjzrT7YXjZvD4PhVzJg7DMpqrm5ji92RGRQwT3LqK357Y"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet BTC mainnet address",
			args:    args{"LTC", "mainnet", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tc56aenn3yvck7d0x6sc0qgxs65c"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet BCH mainnet address",
			args:    args{"LTC", "mainnet", "1BtBojSMWGpp8z4EgrFbd2BZKiThXRYX1e"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet ETH address",
			args:    args{"LTC", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet DASH mainnet address",
			args:    args{"LTC", "mainnet", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet regtest address",
			args:    args{"LTC", "mainnet", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet testnet4 address",
			args:    args{"LTC", "mainnet", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4p5rksqg"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet simnet address",
			args:    args{"LTC", "mainnet", "sltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgs7dcdra"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet random",
			args:    args{"LTC", "mainnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "LTC mainnet empty",
			args:    args{"LTC", "mainnet", ""},
			wantErr: true,
		},

		// LTC regtest
		{
			name:    "LTC regtest P2PKH uncompressed",
			args:    args{"LTC", "regtest", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"},
			wantErr: false,
		},
		{
			name:    "LTC regtest P2PKH compressed",
			args:    args{"LTC", "regtest", "mwt7UCMmfKcCVz6XUQhPpxEMZbs7Rxn5o3"},
			wantErr: false,
		},
		{
			name:    "LTC regtest P2PKH hybrid",
			args:    args{"LTC", "regtest", "mixTqvNvpNcuitucquDsXCWxVD6GgihbXQ"},
			wantErr: false,
		},
		{
			name:    "LTC regtest P2WPKH ",
			args:    args{"LTC", "regtest", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4p5rksqg"},
			wantErr: false,
		},
		{
			name:    "LTC regtest P2SH",
			args:    args{"LTC", "regtest", "2NDwcejJFpwdKnoGTya4gnWS1n2zsGgYfiA"},
			wantErr: false,
		},
		{
			name:    "LTC regtest P2WSH",
			args:    args{"LTC", "regtest", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"},
			wantErr: false,
		},
		{
			name:    "LTC regtest private WIF uncompressed",
			args:    args{"LTC", "regtest", "92Wv2tcSPWaQd6EV1pZDExrdZmhpgWwUg8HW7rJ1NDmxU5GL78f"},
			wantErr: true,
		},
		{
			name:    "LTC regtest private WIF compressed",
			args:    args{"LTC", "regtest", "cRh35LoR4MYTGJ6WecuB2dY4PeKBTR9T9h9qXEa4cfPRpNhP1ZeS"},
			wantErr: true,
		},
		{
			name:    "LTC regtest ETH address",
			args:    args{"LTC", "regtest", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "LTC regtest DASH regtest address",
			args:    args{"LTC", "regtest", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"},
			wantErr: true,
		},
		{
			name:    "LTC regtest mainnet address",
			args:    args{"LTC", "regtest", "LQasau6s59W354CqQk7Hmks5E7XagUMwxj"},
			wantErr: true,
		},
		{
			name:    "LTC regtest simnet address",
			args:    args{"LTC", "regtest", "rrGamoYDzmoN6A69BwkRPMZaqDC1LJC2hZ"},
			wantErr: true,
		},
		{
			name:    "LTC regtest random",
			args:    args{"LTC", "regtest", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "LTC regtest empty",
			args:    args{"LTC", "regtest", ""},
			wantErr: true,
		},
		// false positives
		{
			name:    "LTC regtest BTC regtest address",
			args:    args{"LTC", "regtest", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"},
			wantErr: false,
		},
		{
			name:    "LTC regtest BCH regtest address",
			args:    args{"LTC", "regtest", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"},
			wantErr: false,
		},
		{
			name:    "LTC regtest testnet3 address",
			args:    args{"LTC", "regtest", "mwt7UCMmfKcCVz6XUQhPpxEMZbs7Rxn5o3"},
			wantErr: false,
		},

		// LTC testnet4
		{
			name:    "LTC testnet4 P2PKH uncompressed",
			args:    args{"LTC", "testnet4", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"},
			wantErr: false,
		},
		{
			name:    "LTC testnet4 P2PKH compressed",
			args:    args{"LTC", "testnet4", "mwt7UCMmfKcCVz6XUQhPpxEMZbs7Rxn5o3"},
			wantErr: false,
		},
		{
			name:    "LTC testnet4 P2PKH hybrid",
			args:    args{"LTC", "testnet4", "mixTqvNvpNcuitucquDsXCWxVD6GgihbXQ"},
			wantErr: false,
		},
		{
			name:    "LTC testnet4 P2WPKH ",
			args:    args{"LTC", "testnet4", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4p5rksqg"},
			wantErr: false,
		},
		{
			name:    "LTC testnet4 P2SH",
			args:    args{"LTC", "testnet4", "2NDwcejJFpwdKnoGTya4gnWS1n2zsGgYfiA"},
			wantErr: false,
		},
		{
			name:    "LTC testnet4 P2WSH",
			args:    args{"LTC", "testnet4", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"},
			wantErr: false,
		},
		{
			name:    "LTC testnet4 private WIF uncompressed",
			args:    args{"LTC", "testnet4", "92Wv2tcSPWaQd6EV1pZDExrdZmhpgWwUg8HW7rJ1NDmxU5GL78f"},
			wantErr: true,
		},
		{
			name:    "LTC testnet4 private WIF compressed",
			args:    args{"LTC", "testnet4", "cRh35LoR4MYTGJ6WecuB2dY4PeKBTR9T9h9qXEa4cfPRpNhP1ZeS"},
			wantErr: true,
		},
		{
			name:    "LTC testnet4 ETH testnet3 address",
			args:    args{"LTC", "testnet4", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "LTC testnet4 DASH address",
			args:    args{"LTC", "testnet4", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"},
			wantErr: true,
		},
		{
			name:    "LTC testnet4 mainnet address",
			args:    args{"LTC", "testnet4", "LQasau6s59W354CqQk7Hmks5E7XagUMwxj"},
			wantErr: true,
		},
		{
			name:    "LTC testnet4 simnet address",
			args:    args{"LTC", "testnet4", "rrGamoYDzmoN6A69BwkRPMZaqDC1LJC2hZ"},
			wantErr: true,
		},
		{
			name:    "LTC testnet4 random",
			args:    args{"LTC", "testnet4", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "LTC testnet4 empty",
			args:    args{"LTC", "testnet4", ""},
			wantErr: true,
		},
		// false positive
		{
			name:    "LTC testnet4 BTC testnet3 address",
			args:    args{"LTC", "testnet4", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"},
			wantErr: false,
		},
		{
			name:    "LTC testnet4 BCH testnet3 address",
			args:    args{"LTC", "testnet4", "mwHfttSrSFWz7MztJNCtubUPGAFqRPHUKi"},
			wantErr: false,
		},
		{
			name:    "LTC testnet4 regtest address",
			args:    args{"LTC", "testnet4", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"},
			wantErr: false,
		},

		// LTC simnet
		{
			name:    "LTC simnet P2PKH uncompressed",
			args:    args{"LTC", "simnet", "SUS7ygcGDz9gxaq9iEaPHXzKRC5aWtrEFP"},
			wantErr: true,
		},
		{
			name:    "LTC simnet P2PKH compressed",
			args:    args{"LTC", "simnet", "SdfACz3wafN9FBQNJGi6YwAbMPVqJ8Cc9w"},
			wantErr: true,
		},
		{
			name:    "LTC simnet P2PKH hybrid",
			args:    args{"LTC", "simnet", "SQjWai56jiNrU6DTfmEaFBTCGzizYTGkQd"},
			wantErr: true,
		},
		{
			name:    "LTC simnet P2WPKH ",
			args:    args{"LTC", "simnet", "sltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pydmk0n"},
			wantErr: true,
		},
		{
			name:    "LTC simnet P2SH",
			args:    args{"LTC", "simnet", "rrGamoYDzmoN6A69BwkRPMZaqDC1LJC2hZ"},
			wantErr: true,
		},
		{
			name:    "LTC simnet P2WSH",
			args:    args{"LTC", "simnet", "sltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgs7dcdra"},
			wantErr: true,
		},
		{
			name:    "LTC simnet private WIF uncompressed",
			args:    args{"LTC", "simnet", "4NKSWfDzwNZvt4TT1FHKfPWzGv71KdomJHSYD2gQC8r3siLQiQt"},
			wantErr: true,
		},
		{
			name:    "LTC simnet private WIF compressed",
			args:    args{"LTC", "simnet", "Fs6PbuiNizCLpNFjn633NSPMoFPWRFTQXHZik6QuN9dVgMzx7PiW"},
			wantErr: true,
		},
		{
			name: "LTC simnet BCH testnet3 address",
			args: args{"LTC", "simnet", "mwHfttSrSFWz7MztJNCtubUPGAFqRPHUKi"},
			// Because Bitcoin Cash has the same network parameters.
			wantErr: false,
		},
		{
			name:    "LTC simnet ETH address",
			args:    args{"LTC", "simnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "LTC simnet DASH regtest address",
			args:    args{"LTC", "simnet", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"},
			wantErr: true,
		},
		{
			name:    "LTC simnet mainnet address",
			args:    args{"LTC", "simnet", "LSN5D48waHCYh5jrLwac1euWydDRtB7M3x"},
			wantErr: true,
		},
		{
			name:    "LTC simnet regtest address",
			args:    args{"LTC", "simnet", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"},

			// Because regression network has the same parameters as regtest.
			wantErr: false,
		},
		{
			name: "LTC simnet testnet3 address",
			args: args{"LTC", "simnet", "mwt7UCMmfKcCVz6XUQhPpxEMZbs7Rxn5o3"},

			// Because regression network has the same parameters as testnet.
			wantErr: false,
		},
		{
			name:    "LTC simnet random",
			args:    args{"LTC", "simnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "LTC simnet empty",
			args:    args{"LTC", "simnet", ""},
			wantErr: true,
		},
		{
			name:    "LTC simnet BTC simnet address",
			args:    args{"LTC", "simnet", "ScgVirScTF4ZeAUiNN8yHx7NWqS21eLuzo"},
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
