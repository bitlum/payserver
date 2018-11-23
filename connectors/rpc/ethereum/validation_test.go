package ethereum

import (
	"testing"
)

func TestValidateAddress(t *testing.T) {
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
		{
			name:    "ETH empty net",
			args:    args{"ETH", "", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: false,
		},
		{
			name:    "ETH mainnet",
			args:    args{"ETH", "mainnet", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: false,
		},
		{
			name:    "ETH ropsten",
			args:    args{"ETH", "ropsten", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: false,
		},
		{
			name:    "ETH kovan",
			args:    args{"ETH", "kovan", "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: false,
		},
		{
			name:    "ETH without prefix",
			args:    args{"ETH", "", "de0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: false,
		},
		{
			name:    "ETH invalid",
			args:    args{"ETH", "", "0xdg0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"},
			wantErr: true,
		},
		{
			name:    "ETH BTC mainnet address",
			args:    args{"ETH", "", "bc1qn6f5cd9rpxtgavsxyk7lgyvgn75mj8tc56aenn3yvck7d0x6sc0qgxs65c"},
			wantErr: true,
		},
		{
			name:    "ETH LTC mainnet address",
			args:    args{"ETH", "", "tltc1ql00u9jm8qwzhv4e53hthz34t5744wh4pdkyuny5fp3feklm8cjgscue3nw"},
			wantErr: true,
		},
		{
			name:    "ETH DASH mainnet address",
			args:    args{"ETH", "", "yYQRhUDfzXn4b6acrnenZdDGRTpDUTqmQs"},
			wantErr: true,
		},
		{
			name:    "ETH random",
			args:    args{"ETH", "", "dGj3h7mvUfYuLGX2LoemYxsMyBQo90qQ20"},
			wantErr: true,
		},
		{
			name:    "ETH empty",
			args:    args{"ETH", "", ""},
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
			if err = ValidateAddress(tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
