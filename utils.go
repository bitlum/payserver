package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bitlum/connector/common"
	"github.com/bitlum/connector/common/core"
)

// fileExists reports whether the named file or directory exists.
// This function is taken from https://github.com/btcsuite/btcd
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func doDeposit(engine *core.Engine, payment *common.Payment,
	asset core.AssetType) {
	userID, err := getUserIDFromAccount(payment.Account)
	if err != nil {
		mainLog.Errorf("unable to convert account"+
			"(%v) into user id during transaction "+
			"notification handling: %v", payment.Account, err)
		return
	}

	actionID, err := getActionID(payment.ID)
	if err != nil {
		mainLog.Errorf("unable to get action"+
			"id from payment(%v) during transaction "+
			"notification handling: %v", payment.ID, err)
		return
	}

	// Infinite cycle in the case if service is unavailable.
	// TODO(anddrew.shvv) make if better, probably persistent task queue?
	for {
		req := &core.BalanceUpdateRequest{
			UserID:     userID,
			Asset:      asset,
			ActionType: core.ActionDeposit,
			Change:     payment.Amount.String(),
			ActionID:   actionID,
			Detail: map[string]interface{}{
				"pt":    payment.Type,
				"pid":   payment.ID,
				"paddr": payment.Address,
			},
		}

		resp, err := engine.BalanceUpdate(req)
		if err != nil || resp.Status != "success" {
			mainLog.Errorf("unable to update user balance, "+
				"user(%v), amount(%v), payment(%v), type(%v): %v", payment.Account,
				payment.Amount.String(), payment.ID, payment.Type, err)

			if strings.Contains(err.Error(), "repeat") {
				// If errors contain the repeat it means that for some reason
				// we are trying to deposit using the same payment id.
				return
			} else {
				// If server unable to be reached for some reason,
				// than try it again.
				<-time.After(time.Second)
				continue
			}
		}

		break
	}

	mainLog.Infof("User balance updated: user(%v), amount(%v), payment(%v),"+
		" asset(%v), type(%v)", payment.Account, payment.Amount.String(), payment.ID,
		asset, payment.Type)
	return
}

func getUserIDFromAccount(account string) (uint32, error) {
	id, err := strconv.ParseUint(account, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(id), nil
}

func getActionID(paymentID string) (int32, error) {
	hasher := sha1.New()
	if _, err := hasher.Write([]byte(paymentID)); err != nil {
		return 0, err
	}
	sha := hasher.Sum(nil)

	var actionID int32
	buf := bytes.NewBuffer(sha[:])

	if err := binary.Read(buf, binary.LittleEndian, &actionID); err != nil {
		return 0, err
	}
	return actionID, nil
}
