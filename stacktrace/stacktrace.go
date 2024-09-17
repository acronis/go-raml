package stacktrace

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Type is the type of the error.
type Type string

const (
	TypeUnknown    Type = "unknown"
	TypeParsing    Type = "parsing"
	TypeLoading    Type = "loading"
	TypeReading    Type = "reading"
	TypeResolving  Type = "resolving"
	TypeValidating Type = "validating"
	TypeUnwrapping Type = "unwrapping"
)

// Severity is the severity of the error.
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

// ensureMap ensures that the map is initialized.
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

// StackTrace contains information about a parser error.
type StackTrace struct {
	// Severity is the severity of the error.
	Severity Severity
	// Type is the type of the error.
	Type Type
	// Location is the location file path of the error.
	Location string
	// Position is the position of the error in the file.
	Position *Position

	// Wrapped is the error that wrapped by this error.
	Wrapped *StackTrace
	// Err is the underlying error. It is not used for the error message.
	Err error
	// Message is the error message.
	Message string
	// WrappingMessage is the error messages of the wrapped errors.
	WrappingMessage string
	// Info is the additional information about the error.
	Info StructInfo

	List []*StackTrace

	typeIsSet bool
}

// Header returns the header of the StackTrace.
func (st *StackTrace) Header() string {
	result := fmt.Sprintf("[%s] %s: %s",
		st.Severity,
		st.Type,
		st.Location,
	)
	if st.Position != nil {
		result = fmt.Sprintf("%s:%d:%d", result, st.Position.Line, st.Position.Column)
	} else {
		result = fmt.Sprintf("%s:1", result)
	}
	return result
}

// FullMessage returns the full message of the error including the wrapped messages.
func (st *StackTrace) FullMessage() string {
	if st.WrappingMessage != "" {
		if st.Message != "" {
			return fmt.Sprintf("%s: %s", st.WrappingMessage, st.Message)
		}
		return st.WrappingMessage
	} else {
		return st.Message
	}
}

// Option is an option for the StackTrace creation.
type Option interface {
	Apply(*StackTrace)
}

// OrigString returns the original error message without the wrapping messages.
func (st *StackTrace) OrigString() string {
	result := st.Header()
	if st.Message != "" {
		result = fmt.Sprintf("%s: %s", result, st.Message)
	}
	if len(st.Info.info) > 0 {
		result = fmt.Sprintf("%s: %s", result, st.Info.String())
	}
	return result
}

// OrigStringW returns the original error message with the wrapped error messages
func (st *StackTrace) OrigStringW() string {
	result := st.OrigString()
	if st.Wrapped != nil {
		result = fmt.Sprintf("%s: %s", result, st.Wrapped.String())
	}
	return result
}

// String implements the fmt.Stringer interface.
// It returns the string representation of the StackTrace.
func (st *StackTrace) String() string {
	result := st.Header()
	msg := st.FullMessage()
	if msg != "" {
		result = fmt.Sprintf("%s: %s", result, msg)
	}
	if len(st.Info.info) > 0 {
		result = fmt.Sprintf("%s: %s", result, st.Info.String())
	}
	if st.Wrapped != nil {
		result = fmt.Sprintf("%s: %s", result, st.Wrapped.String())
	}
	if len(st.List) > 0 {
		result = fmt.Sprintf("%s: and more (%d)...", result, len(st.List))
	}
	return result
}

// StackTrace implements the error interface.
// It returns the string representation of the StackTrace.
func (st *StackTrace) Error() string {
	return st.String()
}

// Unwrap checks if the given error is an StackTrace and returns it.
// It returns false if the error is not an StackTrace.
func Unwrap(err error) (*StackTrace, bool) {
	if err == nil {
		return nil, false
	}
	err = FixYamlError(err)
	st, ok := err.(*StackTrace)
	if !ok {
		wrappedErr := errors.Unwrap(err)
		if wrappedErr == nil {
			return nil, false
		}
		st, ok = Unwrap(wrappedErr)
		if ok {
			msg := strings.ReplaceAll(err.Error(), st.OrigStringW(), "")
			msg = strings.TrimSuffix(msg, ": ")
			st.WrappingMessage = msg
			st.Err = err
		}
	}

	// Clone the error to avoid modifying the original error.
	return st.Clone(), ok
}

// New creates a new StackTrace.
func New(message string, location string, opts ...Option) *StackTrace {
	e := &StackTrace{
		Severity: SeverityError,
		Type:     TypeParsing,
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

func (o optErrInfo) Apply(e *StackTrace) {
	e.Info.Add(o.Key, o.Value)
}

type optErrPosition struct {
	Pos *Position
}

func (o optErrPosition) Apply(e *StackTrace) {
	e.Position = o.Pos
}

type optErrSeverity struct {
	Severity Severity
}

func (o optErrSeverity) Apply(e *StackTrace) {
	e.Severity = o.Severity
}

type optErrType struct {
	ErrType Type
}

func (o optErrType) Apply(e *StackTrace) {
	_ = e.SetType(o.ErrType)
}

func WithInfo(key string, value any) Option {
	return optErrInfo{Key: key, Value: Stringer(value)}
}

func WithPosition(pos *Position) Option {
	return optErrPosition{Pos: pos}
}

func WithSeverity(severity Severity) Option {
	return optErrSeverity{Severity: severity}
}

// WithType sets the type of the error with override.
func WithType(errType Type) Option {
	return optErrType{ErrType: errType}
}

// NewWrapped creates a new StackTrace from the given go error.
func NewWrapped(message string, err error, location string, opts ...Option) *StackTrace {
	err = FixYamlError(err)
	resultErr := &StackTrace{}
	if st, ok := Unwrap(err); ok {
		resultErr = New(
			message,
			location,
		).Wrap(st).SetErr(st.Err)
	} else {
		resultErr = New(fmt.Sprintf("%s: %s", message, err.Error()), location).SetErr(err)
	}
	for _, opt := range opts {
		opt.Apply(resultErr)
	}
	return resultErr
}

// SetSeverity sets the severity of the StackTrace and returns it
func (st *StackTrace) SetSeverity(severity Severity) *StackTrace {
	st.Severity = severity
	return st
}

// SetType sets the type of the StackTrace and returns it, operation can be done only once.
func (st *StackTrace) SetType(errType Type) *StackTrace {
	if !st.typeIsSet {
		st.Type = errType
		st.typeIsSet = true
	}
	if st.Wrapped != nil {
		_ = st.Wrapped.SetType(errType)
	}
	return st
}

// SetLocation sets the location of the StackTrace and returns it
func (st *StackTrace) SetLocation(location string) *StackTrace {
	st.Location = location
	return st
}

// SetPosition sets the position of the StackTrace and returns it
func (st *StackTrace) SetPosition(pos *Position) *StackTrace {
	st.Position = pos
	return st
}

// SetWrappingMessage sets a message which wraps the message of StackTrace and returns *StackTrace
func (st *StackTrace) SetWrappingMessage(msg string, a ...any) *StackTrace {
	st.WrappingMessage = fmt.Sprintf(msg, a...)
	return st
}

// SetMessage sets the message of the StackTrace and returns it
func (st *StackTrace) SetMessage(message string, a ...any) *StackTrace {
	st.Message = fmt.Sprintf(message, a...)
	return st
}

// SetErr sets the underlying error of the StackTrace and returns it
func (st *StackTrace) SetErr(err error) *StackTrace {
	st.Err = err
	return st
}

// Wrap wraps the given StackTrace and returns it
func (st *StackTrace) Wrap(w *StackTrace) *StackTrace {
	st.Wrapped = w
	return st
}

// Append adds the given StackTrace to the list of StackTraces and returns it
func (st *StackTrace) Append(e *StackTrace) *StackTrace {
	if st.List == nil {
		st.List = make([]*StackTrace, 0)
	}
	st.List = append(st.List, e)
	return st
}

// Clone returns a clone of the StackTrace.
func (st *StackTrace) Clone() *StackTrace {
	if st == nil {
		return nil
	}
	c := *st
	return &c
}

// Position contains the line and column where the error occurred.
type Position struct {
	Line   int
	Column int
}

// NewNodePosition creates a new position from the given node.
func NewNodePosition(node *yaml.Node) *Position {
	return &Position{Line: node.Line, Column: node.Column}
}

// NewPosition creates a new position with the given line and column.
func NewPosition(line, column int) *Position {
	return &Position{Line: line, Column: column}
}

// optErrNodePosition is an option to set the position of the error to the position of the given node.
type optErrNodePosition struct {
	pos *Position
}

// Apply sets the position of the error to the given position.
// implements Option
func (o optErrNodePosition) Apply(e *StackTrace) {
	e.Position = o.pos
}

// WithNodePosition sets the position of the error to the position of the given node.
func WithNodePosition(node *yaml.Node) Option {
	return optErrNodePosition{pos: NewNodePosition(node)}
}
