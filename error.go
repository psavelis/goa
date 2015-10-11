// Package goa standardizes on structured error responses: a request that fails
// because of an unexpected condition produces a response that contains one or
// more structured error(s). Each error object has three keys: a kind (number),
// a title and a message. The title for a given kind is always the same, the
// intent is to provide a human friendly categorization. The message is specific
// to the error occurrence and provides additional details that often include
// contextual information (name of parameters etc.).
//
// The basic data structure backing errors is TypedError which simply contains
// the kind and message. Multiple errors (not just TypedError instances) can
// be encapsulated in a MultiError. Both TypedError and MultiError implement
// the error interface, the Error methods return valid JSON that can be written
// directly to a response body.
//
// The code generated by goagen calls the helper functions exposed in this file
// when it encounters invalid data (wrong type, doesn't validate etc.) such as
// InvalidParamTypeError, InvalidAttributeTypeError etc. These methods take and
// return an error which is a MultiError that gets built over time. The final
// MultiError object then gets serialized into the response and sent back to
// the client.
package goa

import (
	"bytes"
	"fmt"
	"strings"
)

type (
	// ErrorKind is an enum listing the possible types of errors.
	ErrorKind int

	// TypedError describes an error that can be returned in a HTTP response.
	TypedError struct {
		Kind ErrorKind
		Mesg string
	}

	// MultiError records multiple errors.
	MultiError []error
)

const (
	// ErrInvalidParamType is the error produced by the generated code when
	// a request parameter type does not match the design.
	ErrInvalidParamType = iota + 1

	// ErrMissingParam is the error produced by the generated code when a
	// required request parameter is missing.
	ErrMissingParam

	// ErrInvalidAttributeType is the error produced by the generated
	// code when a data structure attribute type does not match the design
	// definition.
	ErrInvalidAttributeType

	// ErrMissingAttribute is the error produced by the generated
	// code when a data structure attribute required by the design
	// definition is missing.
	ErrMissingAttribute

	// ErrInvalidEnumValue is the error produced by the generated code when
	// a values does not match one of the values listed in the attribute
	// definition as being valid (i.e. not part of the enum).
	ErrInvalidEnumValue

	// ErrMissingHeader is the error produced by the generated code when a
	// required header is missing.
	ErrMissingHeader

	// ErrInvalidFormat is the error produced by the generated code when
	// a value does not match the format specified in the attribute
	// definition.
	ErrInvalidFormat

	// ErrInvalidPattern is the error produced by the generated code when
	// a value does not match the regular expression specified in the
	// attribute definition.
	ErrInvalidPattern

	// ErrInvalidRange is the error produced by the generated code when
	// a value is less than the minimum specified in the design definition
	// or more than the maximum.
	ErrInvalidRange

	// ErrInvalidLength is the error produced by the generated code when
	// a value is a slice with less elements than the minimum length
	// specified in the design definition or more elements than the
	// maximum length.
	ErrInvalidLength
)

// Title returns a human friendly error title
func (k ErrorKind) Title() string {
	switch k {
	case ErrInvalidParamType:
		return "invalid parameter value"
	case ErrMissingParam:
		return "missing required parameter"
	case ErrInvalidAttributeType:
		return "invalid attribute type"
	case ErrMissingAttribute:
		return "missing required attribute"
	case ErrMissingHeader:
		return "missing required HTTP header"
	case ErrInvalidEnumValue:
		return "invalid value"
	case ErrInvalidRange:
		return "invalid value range"
	case ErrInvalidLength:
		return "invalid value length"
	}
	panic("unknown kind")
}

// Error builds an error message from the typed error details.
func (t *TypedError) Error() string {
	var buffer bytes.Buffer
	buffer.WriteString(`{"kind":"`)
	buffer.WriteString(t.Kind.Title())
	buffer.WriteString(`","msg":"`)
	buffer.WriteString(t.Mesg)
	buffer.WriteString(`"}`)
	return buffer.String()
}

// Error summarizes all the underlying error messages in one JSON array.
func (m MultiError) Error() string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	for i, err := range m {
		txt := err.Error()
		if _, ok := err.(*TypedError); !ok {
			txt = fmt.Sprintf(`"%s"`, txt)
		}
		buffer.WriteString(txt)
		if i < len(m)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]")
	return buffer.String()
}

// InvalidParamTypeError appends a typed error of kind ErrInvalidParamType to
// err and returns it.
func InvalidParamTypeError(name string, val interface{}, expected string, err error) error {
	terr := TypedError{
		Kind: ErrInvalidParamType,
		Mesg: fmt.Sprintf("invalid value %#v for parameter %#v, must be a %s",
			val, name, expected),
	}
	return ReportError(err, &terr)
}

// MissingParamError appends a typed error of kind ErrMissingParam to err and
// returns it.
func MissingParamError(name string, err error) error {
	terr := TypedError{
		Kind: ErrMissingParam,
		Mesg: fmt.Sprintf("missing required parameter %#v", name),
	}
	return ReportError(err, &terr)
}

// InvalidAttributeTypeError appends a typed error of kind ErrIncompatibleType
// to err and returns it.
func InvalidAttributeTypeError(ctx string, val interface{}, expected string, err error) error {
	terr := TypedError{
		Kind: ErrInvalidAttributeType,
		Mesg: fmt.Sprintf("type of %s must be %s but got value %#v", ctx,
			expected, val),
	}
	return ReportError(err, &terr)
}

// MissingAttributeError appends a typed error of kind ErrMissingAttribute to
// err and returns it.
func MissingAttributeError(ctx, name string, err error) error {
	terr := TypedError{
		Kind: ErrMissingAttribute,
		Mesg: fmt.Sprintf("attribute %#v of %s is missing and required", name, ctx),
	}
	return ReportError(err, &terr)
}

// MissingHeaderError appends a typed error of kind ErrMissingHeader to err and
// returns it.
func MissingHeaderError(name string, err error) error {
	terr := TypedError{
		Kind: ErrMissingHeader,
		Mesg: fmt.Sprintf("missing required HTTP header %#v", name),
	}
	return ReportError(err, &terr)
}

// InvalidEnumValueError appends a typed error of kind ErrInvalidEnumValue to
// err and returns it.
func InvalidEnumValueError(ctx string, val interface{}, allowed []interface{}, err error) error {
	elems := make([]string, len(allowed))
	for i, a := range allowed {
		elems[i] = fmt.Sprintf("%#v", a)
	}
	terr := TypedError{
		Kind: ErrInvalidEnumValue,
		Mesg: fmt.Sprintf("value of %s must be one of %s but got value %#v", ctx,
			strings.Join(elems, ", "), val),
	}
	return ReportError(err, &terr)
}

// InvalidFormatError appends a typed error of kind ErrInvalidFormat to err and
// returns it.
func InvalidFormatError(ctx, target string, format Format, formatError, err error) error {
	terr := TypedError{
		Kind: ErrInvalidFormat,
		Mesg: fmt.Sprintf("%s must be formatted as a %s but got value %#v, %s",
			ctx, format, target, formatError.Error()),
	}
	return ReportError(err, &terr)
}

// InvalidPatternError appends a typed error of kind ErrInvalidPattern to err and
// returns it.
func InvalidPatternError(ctx, target string, pattern string, err error) error {
	terr := TypedError{
		Kind: ErrInvalidPattern,
		Mesg: fmt.Sprintf("%s must be match the regexp %#v but got value %#v",
			ctx, pattern, target),
	}
	return ReportError(err, &terr)
}

// InvalidRangeError appends a typed error of kind ErrInvalidRange to err and
// returns it.
func InvalidRangeError(ctx, target string, value int, min bool, err error) error {
	comp := "greater or equal to"
	if !min {
		comp = "lesser or equal to"
	}
	terr := TypedError{
		Kind: ErrInvalidRange,
		Mesg: fmt.Sprintf("%s must be %s than %d but got value %#v",
			ctx, comp, value, target),
	}
	return ReportError(err, &terr)
}

// InvalidLengthError appends a typed error of kind ErrInvalidLength to err and
// returns it.
func InvalidLengthError(ctx, target string, value int, min bool, err error) error {
	comp := "greater or equal to"
	if !min {
		comp = "lesser or equal to"
	}
	terr := TypedError{
		Kind: ErrInvalidRange,
		Mesg: fmt.Sprintf("length of %s must be %s than %d but got value %#v",
			ctx, comp, value, target),
	}
	return ReportError(err, &terr)
}

// ReportError coerces the first argument into a MultiError then appends the second argument and
// returns the resulting MultiError.
func ReportError(err error, err2 error) error {
	if err == nil {
		return MultiError{err2}
	}
	if merr, ok := err.(MultiError); ok {
		return append(merr, err2)
	}
	return MultiError{err, err2}
}
