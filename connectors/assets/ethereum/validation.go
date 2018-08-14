package ethereum

import (
	"fmt"
	"regexp"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

var re = regexp.MustCompile(fmt.Sprintf(`(?:0x)?[0-9a-fA-F]{%d}`,
	eth.AddressLength*2))

// ValidateAddress validates Ethereum address.
func ValidateAddress(address string) error {
	if !eth.IsHexAddress(address) || !re.MatchString(address) {
		return errors.Errorf("invalid hex address")
	}
	return nil
}
