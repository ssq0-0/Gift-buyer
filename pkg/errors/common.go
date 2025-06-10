// Package errors provides a centralized error handling system for the gift buying application.
// It defines common error types, utilities for error wrapping and context preservation,
// and standardized error categories for consistent error handling across the application.
//
// The package organizes errors into logical categories:
//   - OS and general errors for system-level issues
//   - HTTP and network errors for communication failures
//   - Service errors for business logic failures
//   - Blockchain and transaction errors for payment processing
//   - Wallet and key management errors for cryptographic operations
//   - Bridge errors for cross-chain operations
//   - CEX errors for exchange interactions
//   - Config errors for configuration management
//
// Usage example:
//
//	if err := someOperation(); err != nil {
//	    return errors.Wrap(errors.ErrOperationFailed, "additional context")
//	}
package errors

import (
	"errors"
	"fmt"
)

var (
	// OS and general errors

	// ErrNotFound indicates that a requested resource was not found.
	// This is typically used when searching for files, database records,
	// or other resources that don't exist.
	ErrNotFound = New("not found")

	// ErrUnsupportedType indicates an operation on an unsupported data type.
	// Used when the application encounters data types it cannot process.
	ErrUnsupportedType = New("unsupported type")

	// ErrExitProgram indicates a normal program termination request.
	// Used to signal clean application shutdown without indicating an error condition.
	ErrExitProgram = New("exit program")

	// ErrValueEmpty indicates that a required value is empty or nil.
	// Used for validation of required fields and parameters.
	ErrValueEmpty = New("value empty")

	// ErrSelection indicates an error in user selection or input.
	// Used when user input is invalid or out of acceptable range.
	ErrSelection = New("selection error")

	// ErrUnexpectedType indicates that a value has an unexpected type.
	// Used during type assertions and interface conversions.
	ErrUnexpectedType = New("unexpected type")

	// HTTP and network errors

	// ErrConnectionFailed indicates a network connection failure.
	// Used when unable to establish network connections to external services.
	ErrConnectionFailed = New("connection failed")

	// ErrRequestFailed indicates a failed HTTP request.
	// Used when HTTP requests return error status codes or fail to complete.
	ErrRequestFailed = New("request failed")

	// ErrResponseParsing indicates failure to parse an HTTP response.
	// Used when response data cannot be decoded or is in unexpected format.
	ErrResponseParsing = New("failed to parse response")

	// ErrStatusCode indicates an unexpected HTTP status code.
	// Used when API responses return unexpected status codes.
	ErrStatusCode = New("unexpected status code")

	// ErrRateLimitReached indicates that an API rate limit has been reached.
	// Used when external services throttle requests due to rate limiting.
	ErrRateLimitReached = New("rate limit reached")

	// Service errors

	// ErrInvalidParams indicates invalid or missing parameters.
	// Used for input validation and parameter checking.
	ErrInvalidParams = New("invalid or missing parameters")

	// ErrNoCreatedValue indicates failure to create a required value.
	// Used when object creation or initialization fails.
	ErrNoCreatedValue = New("no parameter has been created")

	// ErrFailedInit indicates initialization failure.
	// Used when system components fail to initialize properly.
	ErrFailedInit = New("failed to initialize")

	// Blockchain and transaction errors

	// ErrGasEstimation indicates failure to estimate transaction gas.
	// Used in blockchain operations when gas estimation fails.
	ErrGasEstimation = New("failed to estimate gas")

	// ErrTransactionFailed indicates a failed blockchain transaction.
	// Used when blockchain transactions are rejected or fail to execute.
	ErrTransactionFailed = New("transaction failed")

	// ErrTransactionTimeout indicates a transaction timeout.
	// Used when blockchain transactions take too long to confirm.
	ErrTransactionTimeout = New("transaction wait timeout")

	// ErrContractCall indicates a smart contract call failure.
	// Used when smart contract interactions fail or return errors.
	ErrContractCall = New("contract call failed")

	// ErrChainID indicates failure to get blockchain chain ID.
	// Used when unable to determine the blockchain network.
	ErrChainID = New("failed to get chain ID")

	// ErrNonceRetrieval indicates failure to get transaction nonce.
	// Used when unable to retrieve the next transaction nonce.
	ErrNonceRetrieval = New("failed to get nonce")

	// ErrTxSigning indicates transaction signing failure.
	// Used when cryptographic signing of transactions fails.
	ErrTxSigning = New("failed to sign transaction")

	// ErrTxSending indicates transaction broadcast failure.
	// Used when unable to broadcast signed transactions to the network.
	ErrTxSending = New("failed to send transaction")

	// ErrBalanceEstimation indicates failure to estimate balance.
	// Used when unable to retrieve account balances or estimate costs.
	ErrBalanceEstimation = New("failed to estimate balance")

	// Wallet and key management errors

	// ErrEntropyGeneration indicates failure to generate entropy.
	// Used when cryptographic random number generation fails.
	ErrEntropyGeneration = New("failed to generate entropy")

	// ErrMnemonicGeneration indicates failure to generate mnemonic.
	// Used when BIP39 mnemonic phrase generation fails.
	ErrMnemonicGeneration = New("failed to generate mnemonic")

	// ErrKeyDerivation indicates key derivation failure.
	// Used when hierarchical deterministic key derivation fails.
	ErrKeyDerivation = New("failed to derive key")

	// ErrAddressGeneration indicates address generation failure.
	// Used when unable to generate blockchain addresses from keys.
	ErrAddressGeneration = New("failed to generate address")

	// ErrInvalidKeyFormat indicates invalid key format.
	// Used when cryptographic keys are in unexpected or invalid format.
	ErrInvalidKeyFormat = New("invalid key format")

	// ErrInvalidSeedSize indicates invalid seed size.
	// Used when cryptographic seeds don't meet size requirements.
	ErrInvalidSeedSize = New("invalid seed size")

	// Bridge errors

	// ErrBridgeValidation indicates invalid bridge input.
	// Used when cross-chain bridge parameters are invalid.
	ErrBridgeValidation = New("invalid bridge input")

	// ErrQuoteRetrieval indicates failure to get price quote.
	// Used when unable to retrieve pricing information for trades.
	ErrQuoteRetrieval = New("failed to get quote")

	// ErrInvalidQuote indicates invalid quote response.
	// Used when price quotes are malformed or unreasonable.
	ErrInvalidQuote = New("invalid quote response")

	// ErrHighImpact indicates unacceptable price impact.
	// Used when trades would cause excessive price slippage.
	ErrHighImpact = New("price impact too high")

	// CEX errors

	// ErrCexFailed indicates a failed CEX operation.
	// Used when centralized exchange operations fail.
	ErrCexFailed = New("invalid request")

	// ErrTokenNotFound indicates that a token was not found.
	// Used when requested tokens don't exist on the exchange.
	ErrTokenNotFound = New("token not found")

	// ErrPriceEstimation indicates failure to estimate price.
	// Used when unable to get price estimates for trading pairs.
	ErrPriceEstimation = New("failed to estimate price")

	// Config errors

	// ErrConfigRead indicates failed to read config.
	// Used when configuration files cannot be read from disk.
	ErrConfigRead = New("failed to read config")

	// ErrConfigParse indicates failed to parse config.
	// Used when configuration files contain invalid syntax or structure.
	ErrConfigParse = New("failed to parse config")

	// ErrConfigSave indicates failed to save config.
	// Used when unable to write configuration files to disk.
	ErrConfigSave = New("failed to save config")

	// ErrInvalidConfig indicates invalid configuration.
	// Used when configuration values are invalid or inconsistent.
	ErrInvalidConfig = New("invalid configuration")
)

// New creates a new error with the specified message.
// This is a convenience wrapper around the standard errors.New function
// for consistency with the package's error creation patterns.
//
// Parameters:
//   - text: the error message text
//
// Returns:
//   - error: new error instance with the specified message
func New(text string) error {
	return errors.New(text)
}

// Wrap wraps an existing error with additional context information.
// It creates a new error that includes both the original error and
// additional context, enabling error chain analysis and debugging.
//
// If the provided error is nil, Wrap returns nil to maintain
// idiomatic Go error handling patterns.
//
// Parameters:
//   - err: the original error to wrap (can be nil)
//   - context: additional context information to include
//
// Returns:
//   - error: wrapped error with additional context, or nil if err is nil
//
// Example:
//
//	if err := readFile(); err != nil {
//	    return errors.Wrap(err, "failed to read configuration file")
//	}
func Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}
