package utils

import (
	"errors"
)

type ErrorProcessor struct {
	baseErr error
	chain   []error
	length  int
}

func NewErrorProcessor(err error) *ErrorProcessor {
	if err == nil {
		return &ErrorProcessor{
			baseErr: nil,
			chain:   nil,
			length:  0,
		}
	}

	chain := getAllErrorsFromChain(err)
	return &ErrorProcessor{
		baseErr: err,
		chain:   chain,
		length:  len(chain),
	}
}

// GetErrorByLevel retrieves the error at the specified level from the error chain
//
// A level of 0 returns the original error, positive levels traverse outward,
// and negative levels traverse inward
func (p *ErrorProcessor) GetErrorByLevel(level int) error {
	if p.baseErr == nil {
		return nil
	}

	if p.length == 0 {
		return nil
	}

	var index int
	if level >= 0 {
		index = level
	} else {
		index = p.length + level
	}

	if index < 0 || index >= p.length {
		return nil
	}

	return p.chain[index]
}

// getAllErrorsFromChain retrieves all errors in the error chain
func getAllErrorsFromChain(err error) []error {
	if err == nil {
		return nil
	}

	var errorChain []error
	currentErr := err

	for {
		errorChain = append(errorChain, currentErr)
		unwrapped := errors.Unwrap(currentErr)
		if unwrapped == nil {
			break
		}
		currentErr = unwrapped
	}

	return errorChain
}
