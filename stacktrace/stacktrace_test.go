package stacktrace

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestUnwrapError(t *testing.T) {
	type args struct {
		err error
	}
	firstValidErr := New(
		"first validation error",
		"/usr/local/raml.raml",
	).SetPosition(&Position{Line: 10, Column: 1})
	simpleErr := fmt.Errorf("simple error")
	wrappedSimpleErr := fmt.Errorf("wrapped simpleErr: %w", simpleErr)
	wrappedFirstValidErr := fmt.Errorf("wrapped firstValidErr: %w", firstValidErr)
	secondValidErr := NewWrapped(
		"wrapped firstValidErr",
		firstValidErr,
		"/usr/local/raml2.raml",
	).SetPosition(&Position{Line: 20, Column: 2})
	wrappedSecondValidErr := fmt.Errorf("wrapped secondValidErr: %w", secondValidErr)
	tests := []struct {
		name    string
		args    args
		want    *StackTrace
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
			want: &StackTrace{
				Severity:        SeverityError,
				Type:            TypeValidating,
				Message:         firstValidErr.Message,
				Location:        firstValidErr.Location,
				Position:        firstValidErr.Position,
				WrappingMessage: "wrapped firstValidErr",
				Err:             wrappedFirstValidErr,
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
				err: NewWrapped("wrapped", wrappedSimpleErr, "/usr/local/raml3.raml").SetPosition(&Position{Line: 30, Column: 3}),
			},
			want: &StackTrace{
				Severity:        SeverityError,
				Type:            TypeValidating,
				Message:         fmt.Sprintf("wrapped: %s", wrappedSimpleErr.Error()),
				Location:        "/usr/local/raml3.raml",
				Position:        &Position{Line: 30, Column: 3},
				Err:             wrappedSimpleErr,
				WrappingMessage: "",
			},
			want1:   true,
			wantMsg: fmt.Sprintf("[error] validating: /usr/local/raml3.raml:30:3: wrapped: %s", wrappedSimpleErr.Error()),
		},
		{
			name: "Check double wrapped error",
			args: args{
				err: wrappedSecondValidErr,
			},
			want: &StackTrace{
				Severity:        SeverityError,
				Type:            TypeValidating,
				Message:         secondValidErr.Message,
				Location:        secondValidErr.Location,
				Position:        secondValidErr.Position,
				Err:             wrappedSecondValidErr,
				WrappingMessage: "wrapped secondValidErr",
				Wrapped: &StackTrace{
					Severity: SeverityError,
					Type:     TypeValidating,
					Message:  firstValidErr.Message,
					Location: firstValidErr.Location,
					Position: firstValidErr.Position,
				},
			},
			want1:   true,
			wantMsg: "[error] validating: /usr/local/raml2.raml:20:2: wrapped secondValidErr: wrapped firstValidErr: [error] validating: /usr/local/raml.raml:10:1: first validation error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := Unwrap(tt.args.err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unwrap()\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Unwrap() got1 = %v, want %v", got1, tt.want1)
			}
			if got != nil {
				if got.Error() != tt.wantMsg {
					t.Errorf("Unwrap() Message\ngot:\n%s\nwant:\n%s\n", got.Error(), tt.wantMsg)
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

func TestError_SetSeverity(t *testing.T) {
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
		want   *StackTrace
	}{
		{
			name: "Check set severity",
			fields: fields{
				Severity: SeverityError,
			},
			args: args{
				severity: SeverityCritical,
			},
			want: &StackTrace{
				Severity: SeverityCritical,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &StackTrace{
				Severity: tt.fields.Severity,
			}
			if got := v.SetSeverity(tt.args.severity); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_SetType(t *testing.T) {
	type fields struct {
		ErrType Type
	}
	type args struct {
		errType Type
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *StackTrace
	}{
		{
			name: "Check set type",
			fields: fields{
				ErrType: TypeValidating,
			},
			args: args{
				errType: TypeParsing,
			},
			want: &StackTrace{
				Type: TypeParsing,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &StackTrace{
				Type: tt.fields.ErrType,
			}
			if got := v.SetType(tt.args.errType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_SetLocationAndPosition(t *testing.T) {
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
		want   *StackTrace
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
			want: &StackTrace{
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
			v := &StackTrace{
				Location: tt.fields.Location,
				Position: &tt.fields.Position,
			}
			if got := v.SetLocation(tt.args.location).SetPosition(&tt.args.pos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_SetMessage(t *testing.T) {
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
		want   *StackTrace
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
			want: &StackTrace{
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
			want: &StackTrace{
				Message: "new message with argument",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &StackTrace{
				Message: tt.fields.Message,
			}
			if got := v.SetMessage(tt.args.message, tt.args.a...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_SetWrappingMessage(t *testing.T) {
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
		want   *StackTrace
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
			want: &StackTrace{
				WrappingMessage: "new message",
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
			want: &StackTrace{
				WrappingMessage: "new message with argument",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &StackTrace{
				WrappingMessage: tt.fields.Message,
			}
			if got := v.SetWrappingMessage(tt.args.message, tt.args.a...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetWrappingMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	type fields struct {
		Severity        Severity
		ErrType         Type
		Location        string
		Position        *Position
		Wrapped         *StackTrace
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
				ErrType:         TypeValidating,
				Location:        "/usr/local/raml.raml",
				Position:        &Position{Line: 10, Column: 1},
				Message:         "message",
				WrappedMessages: "wrapped message",
			},
			want: "[error] validating: /usr/local/raml.raml:10:1: wrapped message: message",
		},
		{
			name: "Check error without position",
			fields: fields{
				Severity: SeverityError,
				ErrType:  TypeValidating,
				Location: "/usr/local/raml.raml",
				Message:  "message",
			},
			want: "[error] validating: /usr/local/raml.raml:1: message",
		},
		{
			name: "Check error with info",
			fields: fields{
				Severity: SeverityError,
				ErrType:  TypeValidating,
				Location: "/usr/local/raml.raml",
				Position: &Position{Line: 10, Column: 1},
				Message:  "message",
				Info:     *NewStructInfo().Add("key", Stringer("value")),
			},
			want: "[error] validating: /usr/local/raml.raml:10:1: message: key: value",
		},
		{
			name: "Check error with empty message",
			fields: fields{
				Severity: SeverityError,
				ErrType:  TypeValidating,
				Location: "/usr/local/raml.raml",
				Position: &Position{Line: 10, Column: 1},
			},
			want: "[error] validating: /usr/local/raml.raml:10:1",
		},
		{
			name: "Check error with only wrapped messages",
			fields: fields{
				Severity:        SeverityError,
				ErrType:         TypeValidating,
				Location:        "/usr/local/raml.raml",
				Position:        &Position{Line: 10, Column: 1},
				WrappedMessages: "wrapped message",
			},
			want: "[error] validating: /usr/local/raml.raml:10:1: wrapped message",
		},
		{
			name: "Check error with only wrapped messages and info",
			fields: fields{
				Severity:        SeverityError,
				ErrType:         TypeValidating,
				Location:        "/usr/local/raml.raml",
				Position:        &Position{Line: 10, Column: 1},
				WrappedMessages: "wrapped message",
				Info:            *NewStructInfo().Add("key", Stringer("value")),
			},
			want: "[error] validating: /usr/local/raml.raml:10:1: wrapped message: key: value",
		},
		{
			name: "Check error with only wrapped error",
			fields: fields{
				Severity: SeverityError,
				ErrType:  TypeParsing,
				Location: "/usr/local/raml.raml",
				Position: &Position{Line: 10, Column: 1},
				Wrapped: &StackTrace{
					Severity: SeverityCritical,
					Type:     TypeValidating,
					Location: "/usr/local/raml2.raml",
					Position: &Position{Line: 20, Column: 2},
					Message:  "message 1",
					Wrapped: &StackTrace{
						Severity: SeverityError,
						Type:     TypeResolving,
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
			v := &StackTrace{
				Severity:        tt.fields.Severity,
				Type:            tt.fields.ErrType,
				Location:        tt.fields.Location,
				Position:        tt.fields.Position,
				Wrapped:         tt.fields.Wrapped,
				Err:             tt.fields.Err,
				Message:         tt.fields.Message,
				WrappingMessage: tt.fields.WrappedMessages,
				Info:            tt.fields.Info,
			}
			if got := v.Error(); got != tt.want {
				t.Errorf("StackTrace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_OrigString(t *testing.T) {
	type fields struct {
		Severity        Severity
		ErrType         Type
		Location        string
		Position        Position
		Wrapped         *StackTrace
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
				ErrType:  TypeValidating,
				Location: "/usr/local/raml.raml",
				Position: Position{Line: 10, Column: 1},
				Message:  "message",
				Info:     *NewStructInfo().Add("key", Stringer("value")),
				Wrapped: &StackTrace{
					Severity: SeverityError,
					Type:     TypeValidating,
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
			v := &StackTrace{
				Severity:        tt.fields.Severity,
				Type:            tt.fields.ErrType,
				Location:        tt.fields.Location,
				Position:        &tt.fields.Position,
				Wrapped:         tt.fields.Wrapped,
				Err:             tt.fields.Err,
				Message:         tt.fields.Message,
				WrappingMessage: tt.fields.WrappedMessages,
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

func TestNewWrappedError(t *testing.T) {
	type args struct {
		message  string
		err      error
		location string
		opts     []Option
	}
	tests := []struct {
		name string
		args args
		want *StackTrace
	}{
		{
			name: "Check wrapped error",
			args: args{
				message:  "message",
				err:      fmt.Errorf("error"),
				location: "/usr/local/raml.raml",
				opts: []Option{
					WithSeverity(SeverityCritical),
					WithPosition(NewPosition(10, 1)),
					WithInfo("key", Stringer("value")),
					WithType(TypeParsing),
				},
			},
			want: &StackTrace{
				Severity: SeverityCritical,
				Type:     TypeParsing,
				Message:  "message: error",
				Location: "/usr/local/raml.raml",
				Position: &Position{Line: 10, Column: 1},
				Err:      fmt.Errorf("error"),
				Info:     *NewStructInfo().Add("key", Stringer("value")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewWrapped(tt.args.message, tt.args.err, tt.args.location, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWrapped() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	type args struct {
		message  string
		location string
		opts     []Option
	}
	tests := []struct {
		name string
		args args
		want *StackTrace
	}{
		{
			name: "Check error",
			args: args{
				message:  "message",
				location: "/usr/local/raml.raml",
				opts: []Option{
					WithSeverity(SeverityCritical),
					WithPosition(NewPosition(10, 1)),
					WithInfo("key", Stringer("value")),
					WithType(TypeParsing),
				},
			},
			want: &StackTrace{
				Severity: SeverityCritical,
				Type:     TypeParsing,
				Message:  "message",
				Location: "/usr/local/raml.raml",
				Position: &Position{Line: 10, Column: 1},
				Info:     *NewStructInfo().Add("key", Stringer("value")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.message, tt.args.location, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
