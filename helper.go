// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package validator

import (
	"google.golang.org/protobuf/types/known/anypb"
	"strings"
)

type ValidationError struct {
	Field     string
	Violation string
	ErrorMsg  string
	Index     int
	Errors    *ValidationErrors
}

type ValidationErrors struct {
	Errors []*ValidationError
}

func (f *ValidationErrors) IsError() bool {
	return len(f.Errors) > 0
}

func (f *ValidationErrors) Details() []*anypb.Any {
	details := make([]*anypb.Any, 0, len(f.Errors))
	for _, err := range f.Errors {
		details = append(details, &anypb.Any{
			TypeUrl: err.Field,
			Value:   []byte(err.Violation),
		})
	}
	return details
}

func (f *ValidationErrors) AddValidationsError(fieldName, violation string, i int, err interface{}) {
	message := "one or more items failed validation"
	if violation == "message" {
		message = "invalid"
	}

	switch v := err.(type) {
	case nil:
		return
	case *ValidationError:
		v.ErrorMsg = message
		f.Errors = append(f.Errors, v)
	case *ValidationErrors:
		f.Errors = append(f.Errors, &ValidationError{
			Field:     fieldName,
			Violation: violation,
			Index:     i,
			Errors:    v,
			ErrorMsg:  message,
		})
	case string:
		f.Errors = append(f.Errors, &ValidationError{
			Field:     fieldName,
			Violation: violation,
			ErrorMsg:  v,
		})
	case error:
		f.Errors = append(f.Errors, &ValidationError{
			Field:     fieldName,
			Violation: violation,
			ErrorMsg:  v.Error(),
		})
	}
}

func (f *ValidationErrors) AddValidationError(fieldName, violation string, err interface{}) {
	f.AddValidationsError(fieldName, violation, 0, err)
}

func (f *ValidationErrors) Error() string {
	return "bad request"
}

// Validator is a general interface that allows a message to be validated.
type Validator interface {
	Validate() error
}

// Validators is a general interface that allows all message fields to be validated.
type Validators interface {
	ValidateAll() error
}

func CallValidatorIfExists(candidate interface{}) error {
	if validator, ok := candidate.(Validator); ok {
		return validator.Validate()
	}
	return nil
}

func CallValidatorsIfExists(candidate interface{}) error {
	if validator, ok := candidate.(Validators); ok {
		return validator.ValidateAll()
	}
	return nil
}

type fieldError struct {
	fieldStack []string
	nestedErr  error
}

func (f *fieldError) Error() string {
	return "invalid field " + strings.Join(f.fieldStack, ".") + ": " + f.nestedErr.Error()
}

// FieldError wraps a given Validator error providing a message call stack.
func FieldError(fieldName string, err error) error {
	if fErr, ok := err.(*fieldError); ok {
		fErr.fieldStack = append([]string{fieldName}, fErr.fieldStack...)
		return err
	}
	return &fieldError{
		fieldStack: []string{fieldName},
		nestedErr:  err,
	}
}
