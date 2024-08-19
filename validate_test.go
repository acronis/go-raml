package raml

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestIsValidationError(t *testing.T) {
	type args struct {
		err error
	}
	firstValidErr := NewError(
		"first validation error",
		"/usr/local/raml.raml",
	).SetPosition(Position{Line: 10, Column: 1})
	simpleErr := fmt.Errorf("simple error")
	wrappedSimpleErr := fmt.Errorf("wrapped simpleErr: %w", simpleErr)
	wrappedFirstValidErr := fmt.Errorf("wrapped firstValidErr: %w", firstValidErr)
	secondValidErr := NewWrappedError(
		wrappedFirstValidErr,
		"/usr/local/raml2.raml",
	).SetPosition(Position{Line: 20, Column: 2})
	wrappedSecondValidErr := fmt.Errorf("wrapped secondValidErr: %w", secondValidErr)
	tests := []struct {
		name    string
		args    args
		want    *Error
		want1   bool
		wantMsg string
	}{
		{
			name: "Check the same first validation error",
			args: args{
				err: firstValidErr,
			},
			want:    firstValidErr,
			want1:   true,
			wantMsg: "[error] validating: /usr/local/raml.raml:10:1: first validation error",
		},
		{
			name: "Check wrapped first validation error",
			args: args{
				err: wrappedFirstValidErr,
			},
			want: &Error{
				Severity:        SeverityError,
				ErrType:         ErrTypeValidating,
				Message:         firstValidErr.Message,
				Location:        firstValidErr.Location,
				Position:        firstValidErr.Position,
				WrappedMessages: "wrapped firstValidErr",
				Err:             wrappedFirstValidErr,
				Info:            *NewStructInfo(),
			},
			want1:   true,
			wantMsg: "[error] validating: /usr/local/raml.raml:10:1: wrapped firstValidErr: first validation error",
		},
		{
			name: "Check not a validation error",
			args: args{
				err: simpleErr,
			},
			want:  nil,
			want1: false,
		},
		{
			name: "Check not a validation wrapped error",
			args: args{
				err: wrappedSimpleErr,
			},
			want:  nil,
			want1: false,
		},
		{
			name: "Check err is  nil",
			args: args{
				err: nil,
			},
			want:  nil,
			want1: false,
		},
		{
			name: "Check validation error wrapped in another error",
			args: args{
				err: NewWrappedError(wrappedSimpleErr, "/usr/local/raml3.raml").SetPosition(Position{Line: 30, Column: 3}),
			},
			want: &Error{
				Severity:        SeverityError,
				ErrType:         ErrTypeValidating,
				Message:         wrappedSimpleErr.Error(),
				Location:        "/usr/local/raml3.raml",
				Position:        &Position{Line: 30, Column: 3},
				Err:             wrappedSimpleErr,
				WrappedMessages: "",
				Info:            *NewStructInfo(),
			},
			want1:   true,
			wantMsg: fmt.Sprintf("[error] validating: /usr/local/raml3.raml:30:3: %s", wrappedSimpleErr.Error()),
		},
		{
			name: "Check double wrapped error",
			args: args{
				err: wrappedSecondValidErr,
			},
			want: &Error{
				Severity:        SeverityError,
				ErrType:         ErrTypeValidating,
				Message:         secondValidErr.Message,
				Location:        secondValidErr.Location,
				Position:        secondValidErr.Position,
				Err:             wrappedSecondValidErr,
				WrappedMessages: "wrapped secondValidErr",
				Info:            *NewStructInfo(),
				Wrapped: &Error{
					Severity:        SeverityError,
					ErrType:         ErrTypeValidating,
					Message:         firstValidErr.Message,
					Location:        firstValidErr.Location,
					Position:        firstValidErr.Position,
					Err:             wrappedFirstValidErr,
					WrappedMessages: "wrapped firstValidErr",
					Info:            *NewStructInfo(),
				},
			},
			want1:   true,
			wantMsg: "[error] validating: /usr/local/raml2.raml:20:2: wrapped secondValidErr: [error] validating: /usr/local/raml.raml:10:1: wrapped firstValidErr: first validation error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := UnwrapError(tt.args.err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnwrapError()\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("UnwrapError() got1 = %v, want %v", got1, tt.want1)
			}
			if got != nil {
				if got.Error() != tt.wantMsg {
					t.Errorf("UnwrapError() Message\ngot:\n%s\nwant:\n%s\n", got.Error(), tt.wantMsg)
				}
			}
		})
	}
}

func TestStringer(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Check stringer with string",
			args: args{
				v: "string",
			},
			want: "string",
		},
		{
			name: "Check stringer with int",
			args: args{
				v: 10,
			},
			want: "10",
		},
		{
			name: "Check stringer with stringer",
			args: args{
				v: Stringer("stringer"),
			},
			want: "stringer",
		},
		{
			name: "Check stringer with nil",
			args: args{
				v: nil,
			},
			want: "<nil>",
		},
		{
			name: "Check stringer with error",
			args: args{
				v: fmt.Errorf("error"),
			},
			want: "error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Stringer(tt.args.v); !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("Stringer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructInfo_String(t *testing.T) {
	type fields struct {
		info map[string]fmt.Stringer
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Check empty struct info",
			fields: fields{
				info: map[string]fmt.Stringer{},
			},
			want: "",
		},
		{
			name: "Check struct info with one key",
			fields: fields{
				info: map[string]fmt.Stringer{
					"key": Stringer("value"),
				},
			},
			want: "key: value",
		},
		{
			name: "Check struct info with two keys",
			fields: fields{
				info: map[string]fmt.Stringer{
					"key1": Stringer("value1"),
					"key2": Stringer("value2"),
				},
			},
			want: "key1: value1: key2: value2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StructInfo{
				info: tt.fields.info,
			}
			if got := s.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructInfo_Add(t *testing.T) {
	type fields struct {
		info map[string]fmt.Stringer
	}
	type args struct {
		key   string
		value fmt.Stringer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *StructInfo
	}{
		{
			name: "Check add key",
			fields: fields{
				info: map[string]fmt.Stringer{},
			},
			args: args{
				key:   "key",
				value: Stringer("value"),
			},
			want: &StructInfo{
				info: map[string]fmt.Stringer{
					"key": Stringer("value"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StructInfo{
				info: tt.fields.info,
			}
			if got := s.Add(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructInfo_Get(t *testing.T) {
	type fields struct {
		info map[string]fmt.Stringer
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fmt.Stringer
	}{
		{
			name: "Check get key",
			fields: fields{
				info: map[string]fmt.Stringer{
					"key": Stringer("value"),
				},
			},
			args: args{
				key: "key",
			},
			want: Stringer("value"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StructInfo{
				info: tt.fields.info,
			}
			if got := s.Get(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructInfo_StringBy(t *testing.T) {
	type fields struct {
		info map[string]fmt.Stringer
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Check string by key",
			fields: fields{
				info: map[string]fmt.Stringer{
					"key": Stringer("value"),
				},
			},
			args: args{
				key: "key",
			},
			want: "value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StructInfo{
				info: tt.fields.info,
			}
			if got := s.StringBy(tt.args.key); got != tt.want {
				t.Errorf("StringBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructInfo_Remove(t *testing.T) {
	type fields struct {
		info map[string]fmt.Stringer
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *StructInfo
	}{
		{
			name: "Check remove key",
			fields: fields{
				info: map[string]fmt.Stringer{
					"key": Stringer("value"),
				},
			},
			args: args{
				key: "key",
			},
			want: &StructInfo{
				info: map[string]fmt.Stringer{},
			},
		},
		{
			name: "Check remove key from empty struct info",
			fields: fields{
				info: map[string]fmt.Stringer{},
			},
			args: args{
				key: "key",
			},
			want: &StructInfo{
				info: map[string]fmt.Stringer{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StructInfo{
				info: tt.fields.info,
			}
			if got := s.Remove(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructInfo_Has(t *testing.T) {
	type fields struct {
		info map[string]fmt.Stringer
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Check has key",
			fields: fields{
				info: map[string]fmt.Stringer{
					"key": Stringer("value"),
				},
			},
			args: args{
				key: "key",
			},
			want: true,
		},
		{
			name: "Check has key in empty struct info",
			fields: fields{
				info: map[string]fmt.Stringer{},
			},
			args: args{
				key: "key",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StructInfo{
				info: tt.fields.info,
			}
			if got := s.Has(tt.args.key); got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructInfo_Update(t *testing.T) {
	type fields struct {
		info map[string]fmt.Stringer
	}
	type args struct {
		u *StructInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *StructInfo
	}{
		{
			name: "Check update struct info",
			fields: fields{
				info: map[string]fmt.Stringer{
					"key": Stringer("value"),
				},
			},
			args: args{
				u: &StructInfo{
					info: map[string]fmt.Stringer{
						"key2": Stringer("value2"),
					},
				},
			},
			want: &StructInfo{
				info: map[string]fmt.Stringer{
					"key":  Stringer("value"),
					"key2": Stringer("value2"),
				},
			},
		},
		{
			name: "Check update struct info with the same key",
			fields: fields{
				info: map[string]fmt.Stringer{
					"key": Stringer("value"),
				},
			},
			args: args{
				u: &StructInfo{
					info: map[string]fmt.Stringer{
						"key": Stringer("value2"),
					},
				},
			},
			want: &StructInfo{
				info: map[string]fmt.Stringer{
					"key": Stringer("value2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StructInfo{
				info: tt.fields.info,
			}
			if got := s.Update(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_SetSeverity(t *testing.T) {
	type fields struct {
		Severity Severity
	}
	type args struct {
		severity Severity
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Error
	}{
		{
			name: "Check set severity",
			fields: fields{
				Severity: SeverityError,
			},
			args: args{
				severity: SeverityCritical,
			},
			want: &Error{
				Severity: SeverityCritical,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Error{
				Severity: tt.fields.Severity,
			}
			if got := v.SetSeverity(tt.args.severity); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_SetType(t *testing.T) {
	type fields struct {
		ErrType ErrType
	}
	type args struct {
		errType ErrType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Error
	}{
		{
			name: "Check set type",
			fields: fields{
				ErrType: ErrTypeValidating,
			},
			args: args{
				errType: ErrTypeParsing,
			},
			want: &Error{
				ErrType: ErrTypeParsing,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Error{
				ErrType: tt.fields.ErrType,
			}
			if got := v.SetType(tt.args.errType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_SetLocationAndPosition(t *testing.T) {
	type fields struct {
		Location string
		Position Position
	}
	type args struct {
		location string
		pos      Position
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Error
	}{
		{
			name: "Check set location and position",
			fields: fields{
				Location: "/usr/local/raml.raml",
				Position: Position{
					Line:   10,
					Column: 1,
				},
			},
			args: args{
				location: "/usr/local/raml2.raml",
				pos: Position{
					Line:   20,
					Column: 2,
				},
			},
			want: &Error{
				Location: "/usr/local/raml2.raml",
				Position: &Position{
					Line:   20,
					Column: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Error{
				Location: tt.fields.Location,
				Position: &tt.fields.Position,
			}
			if got := v.SetLocation(tt.args.location).SetPosition(tt.args.pos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_SetMessage(t *testing.T) {
	type fields struct {
		Message string
	}
	type args struct {
		message string
		a       []any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Error
	}{
		{
			name: "Check set message",
			fields: fields{
				Message: "message",
			},
			args: args{
				message: "new message",
				a:       []any{},
			},
			want: &Error{
				Message: "new message",
			},
		},
		{
			name: "Check set message with arguments",
			fields: fields{
				Message: "message",
			},
			args: args{
				message: "new message with %s",
				a:       []any{"argument"},
			},
			want: &Error{
				Message: "new message with argument",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Error{
				Message: tt.fields.Message,
			}
			if got := v.SetMessage(tt.args.message, tt.args.a...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_SetWrappedMessages(t *testing.T) {
	type fields struct {
		Message string
	}
	type args struct {
		message string
		a       []any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Error
	}{
		{
			name: "Check set message",
			fields: fields{
				Message: "message",
			},
			args: args{
				message: "new message",
				a:       []any{},
			},
			want: &Error{
				WrappedMessages: "new message",
			},
		},
		{
			name: "Check set message with arguments",
			fields: fields{
				Message: "message",
			},
			args: args{
				message: "new message with %s",
				a:       []any{"argument"},
			},
			want: &Error{
				WrappedMessages: "new message with argument",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Error{
				WrappedMessages: tt.fields.Message,
			}
			if got := v.SetWrappedMessages(tt.args.message, tt.args.a...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetWrappedMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	type fields struct {
		Severity        Severity
		ErrType         ErrType
		Location        string
		Position        Position
		Wrapped         *Error
		Err             error
		Message         string
		WrappedMessages string
		Info            StructInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Check error",
			fields: fields{
				Severity:        SeverityError,
				ErrType:         ErrTypeValidating,
				Location:        "/usr/local/raml.raml",
				Position:        Position{Line: 10, Column: 1},
				Message:         "message",
				WrappedMessages: "wrapped message",
			},
			want: "[error] validating: /usr/local/raml.raml:10:1: wrapped message: message",
		},
		{
			name: "Check error with info",
			fields: fields{
				Severity: SeverityError,
				ErrType:  ErrTypeValidating,
				Location: "/usr/local/raml.raml",
				Position: Position{Line: 10, Column: 1},
				Message:  "message",
				Info:     *NewStructInfo().Add("key", Stringer("value")),
			},
			want: "[error] validating: /usr/local/raml.raml:10:1: message: key: value",
		},
		{
			name: "Check error with empty message",
			fields: fields{
				Severity: SeverityError,
				ErrType:  ErrTypeValidating,
				Location: "/usr/local/raml.raml",
				Position: Position{Line: 10, Column: 1},
			},
			want: "[error] validating: /usr/local/raml.raml:10:1",
		},
		{
			name: "Check error with only wrapped messages",
			fields: fields{
				Severity:        SeverityError,
				ErrType:         ErrTypeValidating,
				Location:        "/usr/local/raml.raml",
				Position:        Position{Line: 10, Column: 1},
				WrappedMessages: "wrapped message",
			},
			want: "[error] validating: /usr/local/raml.raml:10:1: wrapped message",
		},
		{
			name: "Check error with only wrapped messages and info",
			fields: fields{
				Severity:        SeverityError,
				ErrType:         ErrTypeValidating,
				Location:        "/usr/local/raml.raml",
				Position:        Position{Line: 10, Column: 1},
				WrappedMessages: "wrapped message",
				Info:            *NewStructInfo().Add("key", Stringer("value")),
			},
			want: "[error] validating: /usr/local/raml.raml:10:1: wrapped message: key: value",
		},
		{
			name: "Check error with only wrapped error",
			fields: fields{
				Severity: SeverityError,
				ErrType:  ErrTypeParsing,
				Location: "/usr/local/raml.raml",
				Position: Position{Line: 10, Column: 1},
				Wrapped: &Error{
					Severity: SeverityCritical,
					ErrType:  ErrTypeValidating,
					Location: "/usr/local/raml2.raml",
					Position: &Position{Line: 20, Column: 2},
					Message:  "message 1",
					Wrapped: &Error{
						Severity: SeverityError,
						ErrType:  ErrTypeResolving,
						Location: "/usr/local/raml3.raml",
						Position: &Position{Line: 30, Column: 3},
						Message:  "message 2",
					},
				},
				Message: "message",
			},
			want: "[error] parsing: /usr/local/raml.raml:10:1: message: [critical] validating: /usr/local/raml2.raml:20:2: message 1: [error] resolving: /usr/local/raml3.raml:30:3: message 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Error{
				Severity:        tt.fields.Severity,
				ErrType:         tt.fields.ErrType,
				Location:        tt.fields.Location,
				Position:        &tt.fields.Position,
				Wrapped:         tt.fields.Wrapped,
				Err:             tt.fields.Err,
				Message:         tt.fields.Message,
				WrappedMessages: tt.fields.WrappedMessages,
				Info:            tt.fields.Info,
			}
			if got := v.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_OrigString(t *testing.T) {
	type fields struct {
		Severity        Severity
		ErrType         ErrType
		Location        string
		Position        Position
		Wrapped         *Error
		Err             error
		Message         string
		WrappedMessages string
		Info            StructInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Check original string",
			fields: fields{
				Severity: SeverityError,
				ErrType:  ErrTypeValidating,
				Location: "/usr/local/raml.raml",
				Position: Position{Line: 10, Column: 1},
				Message:  "message",
				Info:     *NewStructInfo().Add("key", Stringer("value")),
				Wrapped: &Error{
					Severity: SeverityError,
					ErrType:  ErrTypeValidating,
					Location: "/usr/local/raml2.raml",
					Position: &Position{Line: 20, Column: 2},
					Message:  "wrapped",
					Info:     *NewStructInfo().Add("key", Stringer("value")),
				},
			},
			want: "[error] validating: /usr/local/raml.raml:10:1: message: key: value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Error{
				Severity:        tt.fields.Severity,
				ErrType:         tt.fields.ErrType,
				Location:        tt.fields.Location,
				Position:        &tt.fields.Position,
				Wrapped:         tt.fields.Wrapped,
				Err:             tt.fields.Err,
				Message:         tt.fields.Message,
				WrappedMessages: tt.fields.WrappedMessages,
				Info:            tt.fields.Info,
			}
			if got := v.OrigString(); got != tt.want {
				t.Errorf("OrigString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetYamlError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want *yaml.TypeError
	}{
		{
			name: "Check yaml error",
			args: args{
				err: &yaml.TypeError{
					Errors: []string{"error"},
				},
			},
			want: &yaml.TypeError{
				Errors: []string{"error"},
			},
		},
		{
			name: "Check not yaml error",
			args: args{
				err: fmt.Errorf("error"),
			},
			want: nil,
		},
		{
			name: "Check wrapped yaml error",
			args: args{
				err: fmt.Errorf("wrapped error: %w", &yaml.TypeError{
					Errors: []string{"error"},
				}),
			},
			want: &yaml.TypeError{
				Errors: []string{
					"wrapped error",
					"error",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetYamlError(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetYamlError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFixYamlError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "Check fix yaml error",
			args: args{
				err: &yaml.TypeError{
					Errors: []string{"error"},
				},
			},
			want: errors.New("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FixYamlError(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FixYamlError() = %v, want %v", got, tt.want)
			}
		})
	}
}
