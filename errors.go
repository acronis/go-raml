package raml

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
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
// It is used to store additional information about an error.
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

func (s *StructInfo) ensureMap() {
	if s.info == nil {
		s.info = make(map[string]fmt.Stringer)
	}
}

// Add adds a key-value pair to the struct info.
func (s *StructInfo) Add(key string, value fmt.Stringer) *StructInfo {
	s.ensureMap()
	s.info[key] = value
	return s
}

// Get returns the value of the given key.
func (s *StructInfo) Get(key string) fmt.Stringer {
	s.ensureMap()
	return s.info[key]
}

// StringBy returns the string value of the given key.
func (s *StructInfo) StringBy(key string) string {
	s.ensureMap()
	return s.info[key].String()
}

// Remove removes the given key from the struct info.
func (s *StructInfo) Remove(key string) *StructInfo {
	s.ensureMap()
	delete(s.info, key)
	return s
}

// Has checks if the given key exists in the struct info.
func (s *StructInfo) Has(key string) bool {
	s.ensureMap()
	_, ok := s.info[key]
	return ok
}

// Keys returns the keys of the struct info.
func (s *StructInfo) Keys() []string {
	s.ensureMap()
	result := make([]string, 0, len(s.info))
	for k := range s.info {
		result = append(result, k)
	}
	return result
}

// SortedKeys returns the sorted keys of the struct info.
func (s *StructInfo) SortedKeys() []string {
	s.ensureMap()
	keys := s.Keys()
	sort.Strings(keys)
	return keys
}

// Update updates the struct info with the given struct info.
func (s *StructInfo) Update(u *StructInfo) *StructInfo {
	s.ensureMap()
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

// Error contains information about a parser error.
type Error struct {
	// Severity is the severity of the error.
	Severity Severity
	// ErrType is the type of the error.
	ErrType ErrType
	// Location is the location file path of the error.
	Location string
	// Position is the position of the error in the file.
	Position *Position

	// Wrapped is the error that wrapped by this error.
	Wrapped *Error
	// Err is the underlying error. It is not used for the error message.
	Err error
	// Message is the error message.
	Message string
	// WrappingMessage is the error messages of the wrapped errors.
	WrappingMessage string
	// Info is the additional information about the error.
	Info StructInfo
}

// Header returns the header of the Error.
func (e *Error) Header() string {
	result := fmt.Sprintf("[%s] %s: %s",
		e.Severity,
		e.ErrType,
		e.Location,
	)
	if e.Position != nil {
		result = fmt.Sprintf("%s:%d:%d", result, e.Position.Line, e.Position.Column)
	} else {
		result = fmt.Sprintf("%s:1", result)
	}
	return result
}

// FullMessage returns the full message of the error including the wrapped messages.
func (e *Error) FullMessage() string {
	if e.WrappingMessage != "" {
		if e.Message != "" {
			return fmt.Sprintf("%s: %s", e.WrappingMessage, e.Message)
		}
		return e.WrappingMessage
	} else {
		return e.Message
	}
}

type ErrOpt interface {
	Apply(*Error)
}

// OrigString returns the original error message without the wrapping messages.
func (e *Error) OrigString() string {
	result := e.Header()
	if e.Message != "" {
		result = fmt.Sprintf("%s: %s", result, e.Message)
	}
	if len(e.Info.info) > 0 {
		result = fmt.Sprintf("%s: %s", result, e.Info.String())
	}
	return result
}

// OrigStringW returns the original error message with the wrapped error messages
func (e *Error) OrigStringW() string {
	result := e.OrigString()
	if e.Wrapped != nil {
		result = fmt.Sprintf("%s: %s", result, e.Wrapped.String())
	}
	return result
}

// String implements the fmt.Stringer interface.
// It returns the string representation of the Error.
func (e *Error) String() string {
	result := e.Header()
	msg := e.FullMessage()
	if msg != "" {
		result = fmt.Sprintf("%s: %s", result, msg)
	}
	if len(e.Info.info) > 0 {
		result = fmt.Sprintf("%s: %s", result, e.Info.String())
	}
	if e.Wrapped != nil {
		result = fmt.Sprintf("%s: %s", result, e.Wrapped.String())
	}
	return result
}

// Error implements the error interface.
// It returns the string representation of the Error.
func (e *Error) Error() string {
	return e.String()
}

// UnwrapError checks if the given error is an Error and returns it.
// It returns false if the error is not an Error.
func UnwrapError(err error) (*Error, bool) {
	err = FixYamlError(err)
	e, ok := err.(*Error)
	if !ok {
		wrappedErr := errors.Unwrap(err)
		if wrappedErr == nil {
			return nil, false
		}
		e, ok = UnwrapError(wrappedErr)
		if ok {
			msg := strings.ReplaceAll(err.Error(), e.OrigStringW(), "")
			msg = strings.TrimSuffix(msg, ": ")
			e.WrappingMessage = msg
			e.Err = err
		}
	}

	// Clone the error to avoid modifying the original error.
	return e.Clone(), ok
}

// NewError creates a new Error.
func NewError(message string, location string, opts ...ErrOpt) *Error {
	e := &Error{
		Severity: SeverityError,
		ErrType:  ErrTypeValidating,
		Message:  message,
		Location: location,
	}
	for _, opt := range opts {
		opt.Apply(e)
	}
	return e
}

// GetYamlError returns the yaml type error from the given error.
// nil if the error is not a yaml type error.
func GetYamlError(err error) *yaml.TypeError {
	if yamlError, ok := err.(*yaml.TypeError); ok {
		return yamlError
	}
	wErr := errors.Unwrap(err)
	if wErr == nil {
		return nil
	} else {
		yamlErr := GetYamlError(wErr)
		if yamlErr != nil {
			toAppend := strings.ReplaceAll(err.Error(), yamlErr.Error(), "")
			toAppend = strings.TrimSuffix(toAppend, ": ")
			// insert the error message in the correct order to the first index
			yamlErr.Errors = append([]string{toAppend}, yamlErr.Errors...)
		}
		return yamlErr
	}
}

// FixYamlError fixes the yaml type error from the given error.
func FixYamlError(err error) error {
	if yamlError := GetYamlError(err); yamlError != nil {
		err = fmt.Errorf("%s", strings.Join(yamlError.Errors, ": "))
	}
	return err
}

type optErrInfo struct {
	Key   string
	Value fmt.Stringer
}

func (o optErrInfo) Apply(e *Error) {
	e.Info.Add(o.Key, o.Value)
}

type optErrPosition struct {
	Pos *Position
}

func (o optErrPosition) Apply(e *Error) {
	e.Position = o.Pos
}

type optErrSeverity struct {
	Severity Severity
}

func (o optErrSeverity) Apply(e *Error) {
	e.Severity = o.Severity
}

type optErrType struct {
	ErrType ErrType
}

func (o optErrType) Apply(e *Error) {
	e.ErrType = o.ErrType
}

func WithInfo(key string, value any) ErrOpt {
	return optErrInfo{Key: key, Value: Stringer(value)}
}

func WithPosition(pos *Position) ErrOpt {
	return optErrPosition{Pos: pos}
}

func WithSeverity(severity Severity) ErrOpt {
	return optErrSeverity{Severity: severity}
}

func WithType(errType ErrType) ErrOpt {
	return optErrType{ErrType: errType}
}

// NewWrappedError creates a new Error from the given go error.
func NewWrappedError(message string, err error, location string, opts ...ErrOpt) *Error {
	err = FixYamlError(err)
	resultErr := &Error{}
	if e, ok := UnwrapError(err); ok {
		resultErr = NewError(
			message,
			location,
		).Wrap(e).SetErr(e.Err)
	} else {
		resultErr = NewError(fmt.Sprintf("%s: %s", message, err.Error()), location).SetErr(err)
	}
	for _, opt := range opts {
		opt.Apply(resultErr)
	}
	return resultErr
}

// SetSeverity sets the severity of the Error and returns it
func (e *Error) SetSeverity(severity Severity) *Error {
	e.Severity = severity
	return e
}

// SetType sets the type of the Error and returns it
func (e *Error) SetType(errType ErrType) *Error {
	e.ErrType = errType
	return e
}

// SetLocation sets the location of the Error and returns it
func (e *Error) SetLocation(location string) *Error {
	e.Location = location
	return e
}

// SetPosition sets the position of the Error and returns it
func (e *Error) SetPosition(pos *Position) *Error {
	e.Position = pos
	return e
}

// SetWrappingMessage sets a message which wraps the message of Error and returns *Error
func (e *Error) SetWrappingMessage(msg string, a ...any) *Error {
	e.WrappingMessage = fmt.Sprintf(msg, a...)
	return e
}

// SetMessage sets the message of the Error and returns it
func (e *Error) SetMessage(message string, a ...any) *Error {
	e.Message = fmt.Sprintf(message, a...)
	return e
}

// SetErr sets the underlying error of the Error and returns it
func (e *Error) SetErr(err error) *Error {
	e.Err = err
	return e
}

// Wrap wraps the given Error and returns it
func (e *Error) Wrap(w *Error) *Error {
	e.Wrapped = w
	return e
}

// Clone returns a clone of the Error.
func (e *Error) Clone() *Error {
	if e == nil {
		return nil
	}
	c := *e
	return &c
}
