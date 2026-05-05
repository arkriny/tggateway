// Package tggateway implements client for the
// [Telegram Gateway API](https://core.telegram.org/gateway).
package tggateway

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Client is a Telegram Gateway API client.
type Client struct {
	// Telegram Gateway API token.
	AccessToken string

	// HTTPClient specifies an HTTP client.
	// Defults to http.DefaultClient.
	HTTPClient http.Client
}

// TODO: allow configuring baseURL.
const baseURL = "https://gatewayapi.telegram.org/"

// RequestStatus represents the status of a verification message request.
type RequestStatus struct {
	// Unique identifier of the verification request.
	RequestID string `json:"request_id"`

	// The phone number to which the verification code was sent, in the E.164 format.
	PhoneNumber string `json:"phone_number"`

	// Total request cost incurred by either checkSendAbility or sendVerificationMessage.
	RequestCost float64 `json:"request_cost"`

	// If True, the request fee was refunded.
	IsRefunded *bool `json:"is_refunded"`

	// Remaining balance in credits.
	// Returned only in response to a request that incurs a charge.
	RemainingBalance *float64 `json:"remaining_balance"`

	// The current message delivery status.
	// Returned only if a verification message was sent to the user.
	DeliveryStatus *DeliveryStatus `json:"delivery_status"`

	// The current status of the verification process.
	VerificationStatus *VerificationStatus `json:"verification_status"`

	// Custom payload if it was provided in the request, 0-256 bytes.
	Payload *string `json:"payload"`
}

// Status of the message.
type DeliveryStatusStatus string

const (
	// The message has been sent to the recipient's device(s).
	DeliveryStatusSent DeliveryStatusStatus = "sent"

	// The message has been delivered to the recipient's device(s).
	DeliveryStatusDelivered DeliveryStatusStatus = "delivered"

	// The message has been read by the recipient.
	DeliveryStatusRead DeliveryStatusStatus = "read"

	// The message has expired without being delivered or read.
	DeliveryStatusExpired DeliveryStatusStatus = "expired"

	// The message has been revoked.
	DeliveryStatusRevoked DeliveryStatusStatus = "revoked"
)

// DeliveryStatus represents the delivery status of a message.
type DeliveryStatus struct {
	// The current status of the message.
	Status DeliveryStatusStatus `json:"status"`

	// The timestamp when the status was last updated.
	UpdatedAt int `json:"updated_at"`
}

// Status of the verification process.
type VerificationStatusStatus string

const (
	// The code entered by the user is correct.
	VerificationStatusValid VerificationStatusStatus = "code_valid"

	// The code entered by the user is incorrect.
	VerificationStatusInvalid VerificationStatusStatus = "code_invalid"

	// The maximum number of attempts to enter the code has been exceeded.
	VerificationStatusMaxAttemptsExceeded VerificationStatusStatus = "code_max_attempts_exceeded"

	// The code has expired and can no longer be used for verification.
	VerificationStatusExpired VerificationStatusStatus = "expired"
)

// VerificationStatus represents the verification status of a code.
type VerificationStatus struct {
	// The current status of the verification process.
	Status VerificationStatusStatus `json:"status"`

	// The timestamp for this particular status.
	// Represents the time when the status was last updated.
	UpdatedAt int `json:"updated_at"`

	// The code entered by the user.
	CodeEntered *string `json:"code_entered"`
}

type SendVerificationMessageParams struct {
	// The phone number to which the verification code was sent, in the E.164 format.
	PhoneNumber string `json:"phone_number"`

	// The unique identifier of a previous request from checkSendAbility.
	// If provided, this request will be free of charge.
	RequestID string `json:"request_id,omitempty"`

	// Username of the Telegram channel from which the code will be sent.
	// The specified channel, if any, must be verified and owned by the
	// same account who owns the Gateway API token.
	SenderUsername string `json:"sender_username,omitempty"`

	// The verification code. Use this parameter if you want to set the verification code yourself.
	// Only fully numeric strings between 4 and 8 characters in length are supported.
	// If this parameter is set, code_length is ignored.
	Code string `json:"code,omitempty"`

	// The length of the verification code if Telegram needs to generate it for you.
	// Supported values are from 4 to 8.
	// This is only relevant if you are not using the code parameter to set your own code.
	// Use the checkVerificationStatus method with the code parameter to verify the code entered by the user.
	CodeLength int `json:"code_length,omitempty"`

	// An HTTPS URL where you want to receive delivery reports related to the sent message, 0-256 bytes.
	CallbackURL string `json:"callback_url,omitempty"`

	// Custom payload, 0-128 bytes. This will not be displayed to the user, use it for your internal processes.
	Payload string `json:"payload,omitempty"`

	// Time-to-live (in seconds) before the message expires.
	// If the message is not delivered or read within this time, the
	// request fee will be refunded.
	// Supported values are from 30 to 3600.
	TTL int `json:"ttl,omitempty"`
}

// SendVerificationMessage sends a verification message. Charges will apply
// according to the pricing plan for each successful message delivery.
// Note that this method is always free of charge when used to send codes to
// your own phone number.
func (c *Client) SendVerificationMessage(ctx context.Context, params *SendVerificationMessageParams) (*RequestStatus, error) {
	var result RequestStatus
	if err := c.do(ctx, "sendVerificationMessage", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type CheckSendAbilityParams struct {
	// The phone number for which you want to check our ability to send a
	// verification message, in the E.164 format.
	PhoneNumber string `json:"phone_number"`
}

// CheckSendAbility checks the ability to send a verification
// message to the specified phone number. If the ability to send is confirmed,
// a fee will apply according to the pricing plan. After checking, you can send
// a verification message using the sendVerificationMessage method, providing
// the request_id from this response.
//
// Within the scope of a request_id, only one fee can be charged. Calling
// sendVerificationMessage once with the returned request_id will be free of
// charge, while repeated calls will result in an error. Conversely, calls that
// don't include a request_id will spawn new requests and incur the respective
// fees accordingly.
// Note that this method is always free of charge when used to send codes to
// your own phone number.
//
// In case the message can be sent, returns a RequestStatus object.
// Otherwise, an appropriate error will be returned.
func (c *Client) CheckSendAbility(ctx context.Context, params *CheckSendAbilityParams) (*RequestStatus, error) {
	var result RequestStatus
	if err := c.do(ctx, "checkSendAbility", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type CheckVerificationStatusParams struct {
	// The unique identifier of the verification request whose status you want to check.
	RequestID string `json:"request_id"`

	// The code entered by the user.
	// If provided, the method checks if the code is valid for the relevant request.
	Code string `json:"code,omitempty"`
}

// CheckVerificationStatus checks the status of a verification message that was sent
// previously. If the code was generated by Telegram for you, you can also
// verify the correctness of the code entered by the user using this method.
// Even if you set the code yourself, it is recommended to call this method
// after the user has successfully entered the code, passing the correct code
// in the code parameter, so that Telegram can track the conversion rate of
// your verifications.
func (c *Client) CheckVerificationStatus(ctx context.Context, params *CheckVerificationStatusParams) (*RequestStatus, error) {
	var result RequestStatus
	if err := c.do(ctx, "checkVerificationStatus", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

var ErrSignatureMismatch = errors.New("signature mismatch")

// VerifyReportIntegrity of the delivery report from callback_url.
//
// The Telegram Gateway API can send delivery reports to a user-specified
// callback URL. When you include a callback_url parameter in your request, the
// API will send an HTTP POST request to that URL containing the delivery
// report for the message. The payload of the POST request will be a JSON
// object representing the RequestStatus object.
//
// Your URL must respond with HTTP status code 200 to acknowledge receipt of
// the report. Any other status code will be considered a failure, and the
// service will retry sending the same report up to 10 times with increasing
// delays between attempts. If all retries fail, the report will be considered
// lost.
func (c *Client) VerifyReportIntegrity(r *http.Request) error {
	timestamp := r.Header.Get("X-Request-Timestamp")
	signature := r.Header.Get("X-Request-Signature")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	defer r.Body.Close()

	accessTokenSum := sha256.Sum256([]byte(c.AccessToken))
	mac := hmac.New(sha256.New, accessTokenSum[:])
	mac.Write([]byte(timestamp))
	mac.Write([]byte{'\n'})
	mac.Write(body)
	expectedHex := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedHex), []byte(signature)) {
		return ErrSignatureMismatch
	}
	return nil
}

func (c *Client) do(ctx context.Context, method string, params any, result any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(params); err != nil {
		return fmt.Errorf("encode params: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+method, &buf)
	if err != nil {
		// should never fail
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return r.Unwrap(result)
}
