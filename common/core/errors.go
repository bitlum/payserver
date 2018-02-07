package core

// EngineCodeError the error code which is used to identify the exact problem
// which occurred on the exchange engine side.
type EngineCodeError uint8

const (
	CodeInvalidArgument    EngineCodeError = 1
	CodeInternalError                      = 2
	CodeServiceUnavailable                 = 3
	CodeMethodNotFound                     = 4
	CodeServiceTimeOut                     = 5
	CodeBalanceNotEnough                   = 10
	CodeRepeatUpdate                       = 11
	CodeAmountToSmall                      = 12
	CodeNoEnoughTrader                     = 13
)

type Error struct {
	// Code...
	Code EngineCodeError `json:"code"`

	// Message...
	Message string `json:"message"`
}

// A compile time check to ensure Error implements the error interface.
var _ error = (*Error)(nil)

func (e *Error) Error() string {
	return e.Message
}
