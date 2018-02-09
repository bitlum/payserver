package addr

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
		{"BTC mainnet P2PKH uncompressed", args{"BTC", "mainnet", "1PFMrJdc6K61x945CwA7BAYvtVkNoaPcYx"}, false},
		{"BTC mainnet P2PKH compressed", args{"BTC", "mainnet", "1HDNEqzJWdRkcieyD2AHkPJ2wTDW48BpmM"}, false},
		{"BTC mainnet P2PKH hybrid", args{"BTC", "mainnet", "1k5N27poM1mpZg6ERvpHAYwRZrMwMt8PX"}, false},
		{"BTC mainnet P2WPKH", args{"BTC", "mainnet", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tcnxexy2"}, false},
		{"BTC mainnet P2SH", args{"BTC", "mainnet", "38xPXRp7AZ9XHCnLycRP8rDEeVMG2GYFMg"}, false},
		{"BTC mainnet P2WSH", args{"BTC", "mainnet", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tc56aenn3yvck7d0x6sc0qgxs65c"}, false},
		{"BTC mainnet private WIF uncompressed", args{"BTC", "mainnet", "5J879fJS6etub5VKcR8LW6NLhHoAV7a1z4PU1ut5PTYn7xEYJVs"}, true},
		{"BTC mainnet private WIF compressed", args{"BTC", "mainnet", "KxaN9E3wmxG77Qp19KVP5QhBvq8ks5zjLqxAF2QSN8BG9bxAmyrG"}, true},
		{"BTC mainnet ETH mainnet address", args{"BTC", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"BTC mainnet LTC mainnet address", args{"BTC", "mainnet", "LSN5D48waHCYh5jrLwac1euWydDRtB7M3x"}, true},
		{"BTC mainnet DASH mainnet address", args{"BTC", "mainnet", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"}, true},
		{"BTC mainnet regtest address", args{"BTC", "mainnet", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"}, true},
		{"BTC mainnet testnet3 address", args{"BTC", "mainnet", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"}, true},
		{"BTC mainnet simnet address", args{"BTC", "mainnet", "rY8j22gBpTbVyp17a3F6eJmGzQRQWLmEcK"}, true},
		{"BTC mainnet random", args{"BTC", "mainnet", "iGpzgxRNMhDZ1o5sOPNQ6dHmamEXkrlDiA"}, true},
		{"BTC mainnet empty", args{"BTC", "mainnet", ""}, true},

		// BTC regtest
		{"BTC regtest P2PKH uncompressed", args{"BTC", "regtest", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"}, false},
		{"BTC regtest P2PKH compressed", args{"BTC", "regtest", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"}, false},
		{"BTC regtest P2PKH hybrid", args{"BTC", "regtest", "mgG2f5CocNT2bg9hwzuC75mGHZT4tdguXh"}, false},
		{"BTC regtest P2WPKH", args{"BTC", "regtest", "tb1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tceqz4le"}, false},
		{"BTC regtest P2SH", args{"BTC", "regtest", "2MzWbbAk8n1esUzQtek3FkoCVrqZRj9kPti"}, false},
		{"BTC regtest P2WSH", args{"BTC", "regtest", "tb1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tc56aenn3yvck7d0x6sc0qlwx4wh"}, false},
		{"BTC regtest private WIF uncompressed", args{"BTC", "regtest", "91tjjQ7ygsy3Z8zcEm2FNgvJLx9seH7DL1FR6YEajCHptzT74Ye"}, true},
		{"BTC regtest private WIF compressed", args{"BTC", "regtest", "cNwMc93oD1xNGrHGXjJWSjCFZ4SAXY6RQt6dMSrwsEqGQLx7TAv5"}, true},
		{"BTC regtest ETH address", args{"BTC", "regtest", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"BTC regtest DASH regtest address", args{"BTC", "regtest", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"}, true},
		{"BTC regtest mainnet address", args{"BTC", "regtest", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tcnxexy2"}, true},
		{"BTC regtest simnet address", args{"BTC", "regtest", "rY8j22gBpTbVyp17a3F6eJmGzQRQWLmEcK"}, true},
		{"BTC regtest random", args{"BTC", "regtest", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"BTC regtest empty", args{"BTC", "regtest", ""}, true},
		// false positive
		{"BTC regtest LTC regtest P2PKH address", args{"BTC", "regtest", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"}, false},

		// BTC testnet3
		{"BTC testnet3 P2PKH uncompressed", args{"BTC", "testnet3", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"}, false},
		{"BTC testnet3 P2PKH compressed", args{"BTC", "testnet3", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"}, false},
		{"BTC testnet3 P2PKH hybrid", args{"BTC", "testnet3", "mgG2f5CocNT2bg9hwzuC75mGHZT4tdguXh"}, false},
		{"BTC testnet3 P2WPKH", args{"BTC", "testnet3", "tb1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tceqz4le"}, false},
		{"BTC testnet3 P2SH", args{"BTC", "testnet3", "2MzWbbAk8n1esUzQtek3FkoCVrqZRj9kPti"}, false},
		{"BTC testnet3 P2WSH", args{"BTC", "testnet3", "tb1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tc56aenn3yvck7d0x6sc0qlwx4wh"}, false},
		{"BTC testnet3 private WIF uncompressed", args{"BTC", "testnet3", "91tjjQ7ygsy3Z8zcEm2FNgvJLx9seH7DL1FR6YEajCHptzT74Ye"}, true},
		{"BTC testnet3 private WIF compressed", args{"BTC", "testnet3", "cNwMc93oD1xNGrHGXjJWSjCFZ4SAXY6RQt6dMSrwsEqGQLx7TAv5"}, true},
		{"BTC testnet3 ETH address", args{"BTC", "testnet3", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"BTC testnet3 LTC testnet4 address", args{"BTC", "testnet3", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"}, true},
		{"BTC testnet3 DASH testnet3 address", args{"BTC", "testnet3", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"}, true},
		{"BTC testnet3 mainnet address", args{"BTC", "testnet3", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tcnxexy2"}, true},
		{"BTC testnet3 simnet address", args{"BTC", "testnet3", "rY8j22gBpTbVyp17a3F6eJmGzQRQWLmEcK"}, true},
		{"BTC testnet3 random", args{"BTC", "testnet3", "LZcFOTv5cd0hcMk8vwpK2Mv3kSzRxfjzyT"}, true},
		{"BTC testnet3 empty", args{"BTC", "testnet3", ""}, true},

		// BTC simnet
		{"BTC simnet P2PKH uncompressed", args{"BTC", "simnet", "ScgVirScTF4ZeAUiNN8yHx7NWqS21eLuzo"}, true},
		{"BTC simnet P2PKH compressed", args{"BTC", "simnet", "SSn4xABN9k2yt1HHMZpR5ZSWcNSy7A7bCN"}, true},
		{"BTC simnet P2PKH hybrid", args{"BTC", "simnet", "Si1sRqSYDbZLMWeDKEwEjpcnCkyjNA1itZ"}, true},
		{"BTC simnet P2WPKH ", args{"BTC", "simnet", "sb1q3588f3ckfhshjhraeufhe8t82yhmy5auzzmklx"}, true},
		{"BTC simnet P2SH", args{"BTC", "simnet", "rY8j22gBpTbVyp17a3F6eJmGzQRQWLmEcK"}, true},
		{"BTC simnet P2WSH", args{"BTC", "simnet", "sb1q3588f3ckfhshjhraeufhe8t82yhmy5aukejx7ntmn6jee9usvyaqus5x4x"}, true},
		{"BTC simnet private WIF uncompressed", args{"BTC", "simnet", "4N5H6rneNj8wowMxeghGbBHxdF1UEw5dEoR1Bv6NoB5ThBCUgjs"}, true},
		{"BTC simnet private WIF compressed", args{"BTC", "simnet", "Fr1t4rkTxtz4vAwWdLCjorRrWxq5EVhQwurwqBLjCoUgow4giigq"}, true},
		{"BTC simnet ETH address", args{"BTC", "simnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"BTC simnet DASH mainnet address", args{"BTC", "simnet", "XnmpgX9EYz7zFMf5HwLPXbnv9BKqxunaZW"}, true},
		{"BTC simnet mainnet address", args{"BTC", "simnet", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tcnxexy2"}, true},
		{"BTC simnet regtest address", args{"BTC", "simnet", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"}, true},
		{"BTC simnet testnet3 address", args{"BTC", "simnet", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"}, true},
		{"BTC simnet random", args{"BTC", "simnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"BTC simnet empty", args{"BTC", "simnet", ""}, true},
		{"BTC simnet LTC simnet address", args{"BTC", "simnet", "SdfACz3wafN9FBQNJGi6YwAbMPVqJ8Cc9w"}, true},

		// BCH mainnet
		{"BCH mainnet P2PKH uncompressed", args{"BCH", "mainnet", "1K6aphb1obCKoLSfL7KZyvBS6hogcUzZNy"}, false},
		{"BCH mainnet P2PKH compressed", args{"BCH", "mainnet", "1BtBojSMWGpp8z4EgrFbd2BZKiThXRYX1e"}, false},
		{"BCH mainnet P2PKH hybrid", args{"BCH", "mainnet", "1GmibqMsdE5jLFXGaoEX5gG4QAf8XX4uMZ"}, false},
		{"BCH mainnet P2PKH CashAddr uncompressed", args{"BCH", "mainnet", "bitcoincash:qrrgpy7nffggd9g0fen82lrhtemauurtnuq46g5jl7"}, false},
		{"BCH mainnet P2PKH CashAddr compressed", args{"BCH", "mainnet", "bitcoincash:qpm47l0kukuzjnk2vsp70256s9pd99qs5u2e7gd5f7"}, false},
		{"BCH mainnet P2PKH CashAddr hybrid", args{"BCH", "mainnet", "bitcoincash:qzk0a6jahmytcdqeh56uj6pe036fkcf9rq45kwptuk"}, false},
		{"BCH mainnet P2SH", args{"BCH", "mainnet", "32Y8cHhzt89aZMFKTvDJrfTAdA8VY6rPvp"}, false},
		{"BCH mainnet private WIF uncompressed", args{"BCH", "mainnet", "5KTkH5jkTaDgMfww9R9uUvXtsc8N1rqntvAdVkhXRhoQPtudYuu"}, true},
		{"BCH mainnet private WIF compressed", args{"BCH", "mainnet", "L4V41yugpHcxiEuTz9cJeFHMxN3VknzMa9fW5ussJ35pG4BgjFSf"}, true},
		{"BCH mainnet ETH address", args{"BCH", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"BCH mainnet LTC mainnet address", args{"BCH", "mainnet", "LSN5D48waHCYh5jrLwac1euWydDRtB7M3x"}, true},
		{"BCH mainnet DASH mainnet address", args{"BCH", "mainnet", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"}, true},
		{"BCH mainnet regtest address", args{"BCH", "mainnet", "mycY7kfzccdaaSvH3gHwoqPkxhQPXVzSwz"}, true},
		{"BCH mainnet testnet3 address", args{"BCH", "mainnet", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"}, true},
		{"BCH mainnet random", args{"BCH", "mainnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"BCH mainnet empty", args{"BCH", "mainnet", ""}, true},

		// BCH testnet3
		{"BCH testnet3 P2PKH uncompressed", args{"BCH", "testnet3", "mycY7kfzccdaaSvH3gHwoqPkxhQPXVzSwz"}, false},
		{"BCH testnet3 P2PKH compressed", args{"BCH", "testnet3", "mrQ96nXLKJG4v6XrQRDySwPtBi4QTtRpVU"}, false},
		{"BCH testnet3 P2PKH hybrid", args{"BCH", "testnet3", "mwHfttSrSFWz7MztJNCtubUPGAFqRPHUKi"}, false},
		{"BCH testnet3 P2PKH CashAddr uncompressed", args{"BCH", "testnet3", "bchtest:qrrgpy7nffggd9g0fen82lrhtemauurtnuy870k9cz"}, false},
		{"BCH testnet3 P2PKH CashAddr compressed", args{"BCH", "testnet3", "bchtest:qpm47l0kukuzjnk2vsp70256s9pd99qs5uwt600rwz"}, false},
		{"BCH testnet3 P2PKH CashAddr hybrid", args{"BCH", "testnet3", "bchtest:qzk0a6jahmytcdqeh56uj6pe036fkcf9rq3xjfrum2"}, false},
		{"BCH testnet3 P2SH", args{"BCH", "testnet3", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"}, false},
		{"BCH testnet3 private WIF uncompressed", args{"BCH", "testnet3", "93ENrpZJ3oHpKjTDmm3pMX5rXGV5B2NzEs2aaP42mSYTAwLB77a"}, true},
		{"BCH testnet3 private WIF compressed", args{"BCH", "testnet3", "cUr3UtuYFMKDsgNjNZRS1ZnRabLuRF63eBoyCLLNo9jpWoH6Mxw2"}, true},
		{"BCH testnet3 ETH address", args{"BCH", "testnet3", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"BCH testnet3 LTC testnet4 address", args{"BCH", "testnet3", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"}, true},
		{"BCH testnet3 DASH testnet3 address", args{"BCH", "testnet3", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"}, true},
		{"BCH testnet3 random", args{"BCH", "testnet3", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"BCH testnet3 empty", args{"BCH", "testnet3", ""}, true},

		// BCH regtest
		{"BCH regtest P2PKH uncompressed", args{"BCH", "regtest", "mycY7kfzccdaaSvH3gHwoqPkxhQPXVzSwz"}, false},
		{"BCH regtest P2PKH compressed", args{"BCH", "regtest", "mrQ96nXLKJG4v6XrQRDySwPtBi4QTtRpVU"}, false},
		{"BCH regtest P2PKH hybrid", args{"BCH", "regtest", "mwHfttSrSFWz7MztJNCtubUPGAFqRPHUKi"}, false},
		{"BCH regtest P2PKH CashAddr uncompressed", args{"BCH", "regtest", "bchtest:qrrgpy7nffggd9g0fen82lrhtemauurtnuy870k9cz"}, false},
		{"BCH regtest P2PKH CashAddr compressed", args{"BCH", "regtest", "bchtest:qpm47l0kukuzjnk2vsp70256s9pd99qs5uwt600rwz"}, false},
		{"BCH regtest P2PKH CashAddr hybrid", args{"BCH", "regtest", "bchtest:qzk0a6jahmytcdqeh56uj6pe036fkcf9rq3xjfrum2"}, false},
		{"BCH regtest P2SH", args{"BCH", "regtest", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"}, false},
		{"BCH regtest private WIF uncompressed", args{"BCH", "regtest", "93ENrpZJ3oHpKjTDmm3pMX5rXGV5B2NzEs2aaP42mSYTAwLB77a"}, true},
		{"BCH regtest private WIF compressed", args{"BCH", "regtest", "cUr3UtuYFMKDsgNjNZRS1ZnRabLuRF63eBoyCLLNo9jpWoH6Mxw2"}, true},
		{"BCH regtest ETH address", args{"BCH", "regtest", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"BCH regtest DASH regtest address", args{"BCH", "regtest", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"}, true},
		{"BCH regtest random", args{"BCH", "regtest", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"BCH regtest empty", args{"BCH", "regtest", ""}, true},
		// false positive
		{"BTC regtest LTC mainnet P2PKH address", args{"BTC", "regtest", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"}, false},

		{"ETH empty net", args{"ETH", "", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, false},
		{"ETH mainnet", args{"ETH", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, false},
		{"ETH ropsten", args{"ETH", "ropsten", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, false},
		{"ETH kovan", args{"ETH", "kovan", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, false},
		{"ETH without prefix", args{"ETH", "", "de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, false},
		{"ETH invalid", args{"ETH", "", "0xdg0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"ETH BTC mainnet address", args{"ETH", "", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tc56aenn3yvck7d0x6sc0qgxs65c"}, true},
		{"ETH LTC mainnet address", args{"ETH", "", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"}, true},
		{"ETH DASH mainnet address", args{"ETH", "", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"}, true},
		{"ETH random", args{"ETH", "", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"ETH empty", args{"ETH", "", ""}, true},

		// LTC mainnet
		{"LTC mainnet P2PKH uncompressed", args{"LTC", "mainnet", "LQasau6s59W354CqQk7Hmks5E7XagUMwxj"}, false},
		{"LTC mainnet P2PKH compressed", args{"LTC", "mainnet", "LXQBaiuzH5UqN1P2MSaNxJ4iE1EBUNn19c"}, false},
		{"LTC mainnet P2PKH hybrid", args{"LTC", "mainnet", "LhB4WoN5c9Btx9CRd2eqEsCBbYW6met5mb"}, false},
		{"LTC mainnet P2WPKH ", args{"LTC", "mainnet", "ltc1qupndfjxttgfdtq3k4wzuvyegcdz8uun0t09j0n"}, false},
		{"LTC mainnet P2SH legacy", args{"LTC", "mainnet", "35fs1NJAvtMvL2EsAzFPwtdyEmQk2LTBHs"}, false},
		{"LTC mainnet P2SH", args{"LTC", "mainnet", "ZG26rWenBR7dfXhN1TEJKJ3ySbjVPismuT"}, false},
		{"LTC mainnet P2WSH", args{"LTC", "mainnet", "ltc1qupndfjxttgfdtq3k4wzuvyegcdz8uun0lwa47m8k3e3qvcw4wuusn8p8rk"}, false},
		{"LTC mainnet private WIF uncompressed", args{"LTC", "mainnet", "6vDkyaqgbDCf67Em5jsJzFcZvPucRf2Z7pHvj9UhehmWTqFigeL"}, true},
		{"LTC mainnet private WIF compressed", args{"LTC", "mainnet", "T7uK6kMsjzrT7YXjZvD4PhVzJg7DMpqrm5ji92RGRQwT3LqK357Y"}, true},
		{"LTC mainnet BTC mainnet address", args{"LTC", "mainnet", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tc56aenn3yvck7d0x6sc0qgxs65c"}, true},
		{"LTC mainnet BCH mainnet address", args{"LTC", "mainnet", "1BtBojSMWGpp8z4EgrFbd2BZKiThXRYX1e"}, true},
		{"LTC mainnet ETH address", args{"LTC", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"LTC mainnet DASH mainnet address", args{"LTC", "mainnet", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"}, true},
		{"LTC mainnet regtest address", args{"LTC", "mainnet", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"}, true},
		{"LTC mainnet testnet4 address", args{"LTC", "mainnet", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4p5rksqg"}, true},
		{"LTC mainnet simnet address", args{"LTC", "mainnet", "sltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgs7dcdra"}, true},
		{"LTC mainnet random", args{"LTC", "mainnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"LTC mainnet empty", args{"LTC", "mainnet", ""}, true},

		// LTC regtest
		{"LTC regtest P2PKH uncompressed", args{"LTC", "regtest", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"}, false},
		{"LTC regtest P2PKH compressed", args{"LTC", "regtest", "mwt7UCMmfKcCVz6XUQhPpxEMZbs7Rxn5o3"}, false},
		{"LTC regtest P2PKH hybrid", args{"LTC", "regtest", "mixTqvNvpNcuitucquDsXCWxVD6GgihbXQ"}, false},
		{"LTC regtest P2WPKH ", args{"LTC", "regtest", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4p5rksqg"}, false},
		{"LTC regtest P2SH", args{"LTC", "regtest", "2NDwcejJFpwdKnoGTya4gnWS1n2zsGgYfiA"}, false},
		{"LTC regtest P2WSH", args{"LTC", "regtest", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"}, false},
		{"LTC regtest private WIF uncompressed", args{"LTC", "regtest", "92Wv2tcSPWaQd6EV1pZDExrdZmhpgWwUg8HW7rJ1NDmxU5GL78f"}, true},
		{"LTC regtest private WIF compressed", args{"LTC", "regtest", "cRh35LoR4MYTGJ6WecuB2dY4PeKBTR9T9h9qXEa4cfPRpNhP1ZeS"}, true},
		{"LTC regtest ETH address", args{"LTC", "regtest", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"LTC regtest DASH regtest address", args{"LTC", "regtest", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"}, true},
		{"LTC regtest mainnet address", args{"LTC", "regtest", "LQasau6s59W354CqQk7Hmks5E7XagUMwxj"}, true},
		{"LTC regtest simnet address", args{"LTC", "regtest", "rrGamoYDzmoN6A69BwkRPMZaqDC1LJC2hZ"}, true},
		{"LTC regtest random", args{"LTC", "regtest", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"LTC regtest empty", args{"LTC", "regtest", ""}, true},
		// false positives
		{"LTC regtest BTC regtest address", args{"LTC", "regtest", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"}, false},
		{"LTC regtest BCH regtest address", args{"LTC", "regtest", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"}, false},
		{"LTC regtest testnet3 address", args{"LTC", "regtest", "mwt7UCMmfKcCVz6XUQhPpxEMZbs7Rxn5o3"}, false},

		// LTC testnet4
		{"LTC testnet4 P2PKH uncompressed", args{"LTC", "testnet4", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"}, false},
		{"LTC testnet4 P2PKH compressed", args{"LTC", "testnet4", "mwt7UCMmfKcCVz6XUQhPpxEMZbs7Rxn5o3"}, false},
		{"LTC testnet4 P2PKH hybrid", args{"LTC", "testnet4", "mixTqvNvpNcuitucquDsXCWxVD6GgihbXQ"}, false},
		{"LTC testnet4 P2WPKH ", args{"LTC", "testnet4", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4p5rksqg"}, false},
		{"LTC testnet4 P2SH", args{"LTC", "testnet4", "2NDwcejJFpwdKnoGTya4gnWS1n2zsGgYfiA"}, false},
		{"LTC testnet4 P2WSH", args{"LTC", "testnet4", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"}, false},
		{"LTC testnet4 private WIF uncompressed", args{"LTC", "testnet4", "92Wv2tcSPWaQd6EV1pZDExrdZmhpgWwUg8HW7rJ1NDmxU5GL78f"}, true},
		{"LTC testnet4 private WIF compressed", args{"LTC", "testnet4", "cRh35LoR4MYTGJ6WecuB2dY4PeKBTR9T9h9qXEa4cfPRpNhP1ZeS"}, true},
		{"LTC testnet4 ETH testnet3 address", args{"LTC", "testnet4", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"LTC testnet4 DASH address", args{"LTC", "testnet4", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"}, true},
		{"LTC testnet4 mainnet address", args{"LTC", "testnet4", "LQasau6s59W354CqQk7Hmks5E7XagUMwxj"}, true},
		{"LTC testnet4 simnet address", args{"LTC", "testnet4", "rrGamoYDzmoN6A69BwkRPMZaqDC1LJC2hZ"}, true},
		{"LTC testnet4 random", args{"LTC", "testnet4", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"LTC testnet4 empty", args{"LTC", "testnet4", ""}, true},
		// false positive
		{"LTC testnet4 BTC testnet3 address", args{"LTC", "testnet4", "n3mK9MiauLXGjFXgvW8V15mFkVM5hMXy5V"}, false},
		{"LTC testnet4 BCH testnet3 address", args{"LTC", "testnet4", "mwHfttSrSFWz7MztJNCtubUPGAFqRPHUKi"}, false},
		{"LTC testnet4 regtest address", args{"LTC", "testnet4", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"}, false},

		// LTC simnet
		{"LTC simnet P2PKH uncompressed", args{"LTC", "simnet", "SUS7ygcGDz9gxaq9iEaPHXzKRC5aWtrEFP"}, true},
		{"LTC simnet P2PKH compressed", args{"LTC", "simnet", "SdfACz3wafN9FBQNJGi6YwAbMPVqJ8Cc9w"}, true},
		{"LTC simnet P2PKH hybrid", args{"LTC", "simnet", "SQjWai56jiNrU6DTfmEaFBTCGzizYTGkQd"}, true},
		{"LTC simnet P2WPKH ", args{"LTC", "simnet", "sltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pydmk0n"}, true},
		{"LTC simnet P2SH", args{"LTC", "simnet", "rrGamoYDzmoN6A69BwkRPMZaqDC1LJC2hZ"}, true},
		{"LTC simnet P2WSH", args{"LTC", "simnet", "sltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgs7dcdra"}, true},
		{"LTC simnet private WIF uncompressed", args{"LTC", "simnet", "4NKSWfDzwNZvt4TT1FHKfPWzGv71KdomJHSYD2gQC8r3siLQiQt"}, true},
		{"LTC simnet private WIF compressed", args{"LTC", "simnet", "Fs6PbuiNizCLpNFjn633NSPMoFPWRFTQXHZik6QuN9dVgMzx7PiW"}, true},
		{"LTC simnet BCH testnet3 address", args{"LTC", "simnet", "mwHfttSrSFWz7MztJNCtubUPGAFqRPHUKi"}, true},
		{"LTC simnet ETH address", args{"LTC", "simnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"LTC simnet DASH regtest address", args{"LTC", "simnet", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"}, true},
		{"LTC simnet mainnet address", args{"LTC", "simnet", "LSN5D48waHCYh5jrLwac1euWydDRtB7M3x"}, true},
		{"LTC simnet regtest address", args{"LTC", "simnet", "mnf5Etv6JePkDPXJtNZgZZ45dQSrieJfLk"}, true},
		{"LTC simnet testnet3 address", args{"LTC", "simnet", "mwt7UCMmfKcCVz6XUQhPpxEMZbs7Rxn5o3"}, true},
		{"LTC simnet random", args{"LTC", "simnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"LTC simnet empty", args{"LTC", "simnet", ""}, true},
		{"LTC simnet BTC simnet address", args{"LTC", "simnet", "ScgVirScTF4ZeAUiNN8yHx7NWqS21eLuzo"}, true},

		// DASH mainnet
		{"DASH mainnet P2PKH uncompressed", args{"DASH", "mainnet", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"}, false},
		{"DASH mainnet P2PKH compressed", args{"DASH", "mainnet", "Xm1tMRkcJ7u9c95xYeh7vcGj7jrp1EDr54"}, false},
		{"DASH mainnet P2PKH hybrid", args{"DASH", "mainnet", "XnmpgX9EYz7zFMf5HwLPXbnv9BKqxunaZW"}, false},
		{"DASH mainnet P2SH", args{"DASH", "mainnet", "7fxExScCeJyW6wmQTu8hxPwWk81dp7LJGf"}, false},
		{"DASH mainnet private WIF", args{"DASH", "mainnet", "7sL7xs65eEjDatRMr9YusjpFLXoJYjTfjiXqxhXx5ZpfnLdFRAS"}, true},
		{"DASH mainnet private WIF", args{"DASH", "mainnet", "XK9Pja5RVMbLYZxX1vjrNwyYHCwNT4QhyYK96quc79sQAtJBiPve"}, true},
		{"DASH mainnet BTC mainnet address", args{"DASH", "mainnet", "1HDNEqzJWdRkcieyD2AHkPJ2wTDW48BpmM"}, true},
		{"DASH mainnet BCH mainnet address", args{"DASH", "mainnet", "32Y8cHhzt89aZMFKTvDJrfTAdA8VY6rPvp"}, true},
		{"DASH mainnet ETH address", args{"DASH", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"DASH mainnet LTC mainnet address", args{"DASH", "mainnet", "LNfTp5bn61RiCb8AJUEnyJNPqRrqtPAogm"}, true},
		{"DASH mainnet regtest address", args{"DASH", "mainnet", "yhABgLTC8zqV4ABRTz9xkMnb4A15b8zc57"}, true},
		{"DASH mainnet testnet3 address", args{"DASH", "mainnet", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"}, true},
		{"DASH mainnet random", args{"DASH", "mainnet", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"DASH mainnet empty", args{"DASH", "mainnet", ""}, true},

		// DASH regtest
		{"DASH regtest P2PKH uncompressed", args{"DASH", "regtest", "yhABgLTC8zqV4ABRTz9xkMnb4A15b8zc57"}, false},
		{"DASH regtest P2PKH compressed", args{"DASH", "regtest", "yWeVNNq3jfZDwt1W7W1Wxdh5Q2MBYQSjwc"}, false},
		{"DASH regtest P2PKH hybrid", args{"DASH", "regtest", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"}, false},
		{"DASH regtest P2SH", args{"DASH", "regtest", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"}, false},
		{"DASH regtest private WIF", args{"DASH", "regtest", "93NB8yoCU8Q9Jbm3aSGtFxoMPXMSJNxKH61cVwcKsW4F46H4p1S"}, true},
		{"DASH regtest private WIF", args{"DASH", "regtest", "cVSTkDgucjf9egRQNaZ7F3HazQyCfhu9g17hgk2vHuFK2UJvDmFw"}, true},
		{"DASH regtest BTC regtest address", args{"DASH", "regtest", "mwjKXu5HKes1Pq8avb8faJWMoSpD3TmeCP"}, true},
		{"DASH regtest BCH regtest address", args{"DASH", "regtest", "mrQ96nXLKJG4v6XrQRDySwPtBi4QTtRpVU"}, true},
		{"DASH regtest ETH address", args{"DASH", "regtest", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"DASH regtest LTC regtest address", args{"DASH", "regtest", "mixTqvNvpNcuitucquDsXCWxVD6GgihbXQ"}, true},
		{"DASH regtest mainnet address", args{"DASH", "regtest", "Xm1tMRkcJ7u9c95xYeh7vcGj7jrp1EDr54"}, true},
		{"DASH regtest random", args{"DASH", "regtest", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"DASH regtest empty", args{"DASH", "regtest", ""}, true},
		// false positive
		{"DASH regtest testnet3 address", args{"DASH", "regtest", "yhABgLTC8zqV4ABRTz9xkMnb4A15b8zc57"}, false},

		// DASH testnet3
		{"DASH testnet3 P2PKH uncompressed", args{"DASH", "testnet3", "yhABgLTC8zqV4ABRTz9xkMnb4A15b8zc57"}, false},
		{"DASH testnet3 P2PKH compressed", args{"DASH", "testnet3", "yWeVNNq3jfZDwt1W7W1Wxdh5Q2MBYQSjwc"}, false},
		{"DASH testnet3 P2PKH hybrid", args{"DASH", "testnet3", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"}, false},
		{"DASH testnet3 P2SH", args{"DASH", "testnet3", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"}, false},
		{"DASH testnet3 private WIF", args{"DASH", "testnet3", "93NB8yoCU8Q9Jbm3aSGtFxoMPXMSJNxKH61cVwcKsW4F46H4p1S"}, true},
		{"DASH testnet3 private WIF", args{"DASH", "testnet3", "cVSTkDgucjf9egRQNaZ7F3HazQyCfhu9g17hgk2vHuFK2UJvDmFw"}, true},
		{"DASH testnet3 BTC testnet3 address", args{"DASH", "testnet3", "mgG2f5CocNT2bg9hwzuC75mGHZT4tdguXh"}, true},
		{"DASH testnet3 BCH testnet3 address", args{"DASH", "testnet3", "2Mt6Lg2e2Vaevm8ss93qBUcSRqWLfM8anno"}, true},
		{"DASH testnet3 ETH address", args{"DASH", "testnet3", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"}, true},
		{"DASH testnet3 LTC testnet4 address", args{"DASH", "testnet3", "mixTqvNvpNcuitucquDsXCWxVD6GgihbXQ"}, true},
		{"DASH testnet3 mainnet address", args{"DASH", "testnet3", "XwXafPNkhTBQiRFsu8qZiLNEmsWi9nbTfw"}, true},
		{"DASH testnet3 random", args{"DASH", "testnet3", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"}, true},
		{"DASH testnet3 empty", args{"DASH", "testnet3", ""}, true},
		// false positive
		{"DASH testnet3 regtest address", args{"DASH", "testnet3", "8sy3umW4mrN8ZFBfYA8fQmksddnTwc7niw"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("validateAddr() unexpected panic: %v", r)
				}
			}()
			var err error
			if err = Validate(tt.args.asset, tt.args.net, tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("validateAddr() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
