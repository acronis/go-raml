package raml

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ErrType string

const (
	ErrTypeUnknown    ErrType = "unknown"
	ErrTypeParsing    ErrType = "parsing"
	ErrTypeLoading    ErrType = "loading"
	ErrTypeReading    ErrType = "reading"
	ErrTypeResolving  ErrType = "resolving"
	ErrTypeValidating ErrType = "validating"
)

type Severity string

const (
	SeverityError    Severity = "error"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// stringer is a fmt.Stringer implementation.
type stringer struct {
	msg string
}

// String implements the fmt.Stringer interface.
func (s *stringer) String() string {
	return s.msg
}

// Stringer returns a fmt.Stringer for the given value.
func Stringer(v interface{}) fmt.Stringer {
	switch w := v.(type) {
	case fmt.Stringer:
		return w
	case string:
		return &stringer{msg: w}
	case error:
		return &stringer{msg: w.Error()}
	default:
		return &stringer{msg: fmt.Sprintf("%v", w)}
	}
}

// StructInfo is a map of string keys to fmt.Stringer values.
// It is used to store additional information about a validation error.
// WARNING: Not thread-safe
type StructInfo struct {
	info map[string]fmt.Stringer
}

// String implements the fmt.Stringer interface.
// It returns a string representation of the struct info.
func (s *StructInfo) String() string {
	var result string
	keys := s.SortedKeys()

	for _, k := range keys {
		v, ok := s.info[k]
		if ok {
			if result == "" {
				result = fmt.Sprintf("%s: %s", k, v)
			} else {
				result = fmt.Sprintf("%s: %s: %s", result, k, v)
			}
		}
	}
	return result
}

// Add adds a key-value pair to the struct info.
func (s *StructInfo) Add(key string, value fmt.Stringer) *StructInfo {
	s.info[key] = value
	return s
}

// Get returns the value of the given key.
func (s *StructInfo) Get(key string) fmt.Stringer {
	return s.info[key]
}

// StringBy returns the string value of the given key.
func (s *StructInfo) StringBy(key string) string {
	return s.info[key].String()
}

// Remove removes the given key from the struct info.
func (s *StructInfo) Remove(key string) *StructInfo {
	delete(s.info, key)
	return s
}

// Has checks if the given key exists in the struct info.
func (s *StructInfo) Has(key string) bool {
	_, ok := s.info[key]
	return ok
}

// Keys returns the keys of the struct info.
func (s *StructInfo) Keys() []string {
	result := make([]string, 0, len(s.info))
	for k := range s.info {
		result = append(result, k)
	}
	return result
}

// SortedKeys returns the sorted keys of the struct info.
func (s *StructInfo) SortedKeys() []string {
	keys := s.Keys()
	sort.Strings(keys)
	return keys
}

// Update updates the struct info with the given struct info.
func (s *StructInfo) Update(u *StructInfo) *StructInfo {
	for k, v := range u.info {
		s.info[k] = v
	}
	return s
}

// NewStructInfo creates a new struct info.
func NewStructInfo() *StructInfo {
	return &StructInfo{
		info: make(map[string]fmt.Stringer),
	}
}

// ValidationError contains information about a validation error.
type ValidationError struct {
	// Severity is the severity of the error.
	Severity Severity
	// ErrType is the type of the error.
	ErrType ErrType
	// Location is the location file path of the error.
	Location string
	// Position is the position of the error in the file.
	Position Position

	// Wrapped errors is the validation error that wraps this error.
	Wrapped *ValidationError
	// Err is the underlying error. It is not used for the error message.
	Err error
	// Message is the error message.
	Message string
	// WrappedMessages is the error messages of the wrapped errors.
	WrappedMessages string
	// Info is the additional information about the error.
	Info StructInfo
}

// Header returns the header of the error.
func (v *ValidationError) Header() string {
	return fmt.Sprintf("[%s] %s: %s:%d:%d",
		v.Severity,
		v.ErrType,
		v.Location,
		v.Position.Line,
		v.Position.Column,
	)
}

// FullMessage returns the full message of the error including the wrapped messages.
func (v *ValidationError) FullMessage() string {
	if v.WrappedMessages != "" {
		if v.Message != "" {
			return fmt.Sprintf("%s: %s", v.WrappedMessages, v.Message)
		}
		return v.WrappedMessages
	} else {
		return v.Message
	}
}

// OrigString returns the original error message without the wrapped messages.
func (v *ValidationError) OrigString() string {
	result := v.Header()
	if v.Message != "" {
		result = fmt.Sprintf("%s: %s", result, v.Message)
	}
	if len(v.Info.info) > 0 {
		result = fmt.Sprintf("%s: %s", result, v.Info.String())
	}
	return result
}

// OrigStringW returns the original error message with the wrapped error messages
func (v *ValidationError) OrigStringW() string {
	result := v.OrigString()
	if v.Wrapped != nil {
		result = fmt.Sprintf("%s: %s", result, v.Wrapped.String())
	}
	return result
}

// String implements the fmt.Stringer interface.
func (v *ValidationError) String() string {
	result := v.Header()
	msg := v.FullMessage()
	if msg != "" {
		result = fmt.Sprintf("%s: %s", result, msg)
	}
	if len(v.Info.info) > 0 {
		result = fmt.Sprintf("%s: %s", result, v.Info.String())
	}
	if v.Wrapped != nil {
		result = fmt.Sprintf("%s: %s", result, v.Wrapped.String())
	}
	return result
}

// Error implements the error interface.
func (v *ValidationError) Error() string {
	return v.String()
}

// IsValidationError checks if the given error is a validation error and returns it.
func IsValidationError(err error) (*ValidationError, bool) {
	validationError, ok := err.(*ValidationError)
	if !ok {
		wrappedErr := errors.Unwrap(err)
		if wrappedErr == nil {
			return nil, false
		}
		validationError, ok = IsValidationError(wrappedErr)
		if ok {
			msg := strings.ReplaceAll(err.Error(), validationError.OrigStringW(), "")
			msg = strings.TrimSuffix(msg, ": ")
			validationError.WrappedMessages = msg
			validationError.Err = err
		}
	}

	// Clone the validation error to avoid modifying the original error.
	return validationError.Clone(), ok
}

// NewValidationError creates a new validation error.
func NewValidationError(message string, location string, pos Position) *ValidationError {
	validationError := &ValidationError{
		Severity: SeverityError,
		ErrType:  ErrTypeValidating,
		Message:  message,
		Info:     *NewStructInfo(),
		Location: location,
		Position: pos,
	}
	return validationError
}

// NewValidationErrorFromError creates a new validation error from the given error.
func NewValidationErrorFromError(err error, location string, pos Position) *ValidationError {
	if validationError, ok := IsValidationError(err); ok {
		return NewValidationError(
			"",
			location,
			pos,
		).Wrap(validationError).SetErr(validationError.Err)
	}
	return NewValidationError(err.Error(), location, pos).SetErr(err)
}

// SetSeverity sets the severity of the validation error and returns it
func (v *ValidationError) SetSeverity(severity Severity) *ValidationError {
	v.Severity = severity
	return v
}

// SetType sets the type of the validation error and returns it
func (v *ValidationError) SetType(errType ErrType) *ValidationError {
	v.ErrType = errType
	return v
}

// SetLocation sets the location of the validation error and returns it
func (v *ValidationError) SetLocation(location string) *ValidationError {
	v.Location = location
	return v
}

// SetPosition sets the position of the validation error and returns it
func (v *ValidationError) SetPosition(pos Position) *ValidationError {
	v.Position = pos
	return v
}

// SetWrappedMessages sets the wrapped messages of the validation error and returns it
func (v *ValidationError) SetWrappedMessages(wrappedMessages string, a ...any) *ValidationError {
	v.WrappedMessages = fmt.Sprintf(wrappedMessages, a...)
	return v
}

// SetMessage sets the message of the validation error and returns it
func (v *ValidationError) SetMessage(message string, a ...any) *ValidationError {
	v.Message = fmt.Sprintf(message, a...)
	return v
}

// SetErr sets the underlying error of the validation error and returns it
func (v *ValidationError) SetErr(err error) *ValidationError {
	v.Err = err
	return v
}

// Wrap wraps the given validation error and returns it
func (v *ValidationError) Wrap(w *ValidationError) *ValidationError {
	v.Wrapped = w
	return v
}

// Clone returns a clone of the validation error.
func (v *ValidationError) Clone() *ValidationError {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}
