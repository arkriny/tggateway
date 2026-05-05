package tggateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// response represents top-level JSON object returned by the Telegram Gateway API.
//
// If `Ok` equals true, the request was successful, and the result of the query
// can be found in the `Result` field. In case of an unsuccessful request, `Ok`
// equals false, and the error is explained in the `Error` field.
type response struct {
	Ok     bool            `json:"ok"`
	Error  string          `json:"error"`
	Result json.RawMessage `json:"result"`
}

// Unwrap unmarshals result into the provided pointer or returns a mapped error.
func (r *response) Unwrap(result any) error {
	if !r.Ok {
		// Handle "FLOOD_WAIT_4" and friends.
		if strings.HasPrefix(r.Error, "FLOOD_WAIT_") {
			return ErrFloodWait
		}
		err, found := errMap[r.Error]
		if !found {
			return fmt.Errorf("unknown error: %s", r.Error)
		}
		return err
	}
	if err := json.Unmarshal(r.Result, result); err != nil {
		return fmt.Errorf("unmarshal response result: %w", err)
	}
	return nil
}

var (
	ErrCodeInvalid             = errors.New("code invalid")
	ErrCodeExpired             = errors.New("code expired")
	ErrCodeLengthRequired      = errors.New("code length required")
	ErrCodeLengthInvalid       = errors.New("code length invalid")
	ErrCodeMaxAttemptsExceeded = errors.New("code max attempts exceeded")
	ErrPhoneNumberInvalid      = errors.New("phone number invalid")
	ErrPhoneNumberMismatch     = errors.New("phone number mismatch")
	ErrRequestIDInvalid        = errors.New("request id invalid")
	ErrRequestIDRequired       = errors.New("request id required")
	ErrPayloadInvalid          = errors.New("payload invalid")
	ErrSenderUsernameInvalid   = errors.New("sender username invalid")
	ErrSenderNotVerified       = errors.New("sender not verified")
	ErrSenderNotOwned          = errors.New("sender not owned")
	ErrCallbackURLInvalid      = errors.New("callback URL invalid")
	ErrTTLInvalid              = errors.New("TTL invalid")
	ErrAccessTokenInvalid      = errors.New("access token invalid")
	ErrAccessTokenRequired     = errors.New("access token required")
	ErrMessageAlreadySent      = errors.New("message already sent")
	ErrBalanceNotEnough        = errors.New("balance not enough")
	ErrFloodWait               = errors.New("flood wait")
	ErrUnknownMethod           = errors.New("unknown method")
)

var errMap = map[string]error{
	"CODE_INVALID":               ErrCodeInvalid,
	"CODE_EXPIRED":               ErrCodeExpired,
	"CODE_LENGTH_REQUIRED":       ErrCodeLengthRequired,
	"CODE_LENGTH_INVALID":        ErrCodeLengthInvalid,
	"CODE_MAX_ATTEMPTS_EXCEEDED": ErrCodeMaxAttemptsExceeded,
	"PHONE_NUMBER_INVALID":       ErrPhoneNumberInvalid,
	"PHONE_NUMBER_MISMATCH":      ErrPhoneNumberMismatch,
	"REQUEST_ID_INVALID":         ErrRequestIDInvalid,
	"REQUEST_ID_REQUIRED":        ErrRequestIDRequired,
	"PAYLOAD_INVALID":            ErrPayloadInvalid,
	"SENDER_USERNAME_INVALID":    ErrSenderUsernameInvalid,
	"SENDER_NOT_VERIFIED":        ErrSenderNotVerified,
	"SENDER_NOT_OWNED":           ErrSenderNotOwned,
	"CALLBACK_URL_INVALID":       ErrCallbackURLInvalid,
	"TTL_INVALID":                ErrTTLInvalid,
	"ACCESS_TOKEN_INVALID":       ErrAccessTokenInvalid,
	"ACCESS_TOKEN_REQUIRED":      ErrAccessTokenRequired,
	"MESSAGE_ALREADY_SENT":       ErrMessageAlreadySent,
	"BALANCE_NOT_ENOUGH":         ErrBalanceNotEnough,
	"UNKNOWN_METHOD":             ErrUnknownMethod,
}
