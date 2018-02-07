package core

import (
	"strconv"
	"strings"
	"time"

	"fmt"

	"encoding/json"

	"github.com/pkg/errors"
)

func (t UnixTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

func (t *UnixTime) UnmarshalJSON(s []byte) (err error) {
	parts := strings.Split(string(s), ".")
	if len(parts) != 2 {
		return errors.New("unable to parse time, number should consist of " +
			"two parts")
	}

	seconds, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return err
	}

	nanoseconds, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}

	*(*time.Time)(t) = time.Unix(seconds, nanoseconds)
	return
}

func (t UnixTime) String() string {
	return time.Time(t).String()
}

func (k *Kline) UnmarshalJSON(s []byte) (err error) {
	var values []interface{}
	if err := json.Unmarshal(s, &values); err != nil {
		return err
	}

	if len(values) != 8 {
		return errors.New("unable to decode kline, wrong elements number")
	}

	k.Time = values[0].(float64)
	k.OpenPrice = values[1].(string)
	k.ClosePrice = values[2].(string)
	k.HighestPrice = values[3].(string)
	k.LowestPrice = values[4].(string)
	k.Volume = values[5].(string)
	k.Amount = values[6].(string)
	k.Market = NewMarket(values[7].(string))
	return nil
}

func (a *Depth) UnmarshalJSON(s []byte) (err error) {
	values := make([]string, 0)
	if err := json.Unmarshal(s, &values); err != nil {
		return err
	}

	if len(values) != 2 {
		return errors.New("unable to decode price and volume")
	}

	a.Price = values[0]
	a.Volume = values[1]
	return nil
}

func (a Depth) String() string {
	return fmt.Sprintf("\n\tVolume: %v \n\tPrice: %v", a.Volume, a.Price)
}
