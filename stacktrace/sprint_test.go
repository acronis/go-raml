package stacktrace

import "testing"

func TestError_Sprintf(t *testing.T) {
	type fields struct {
		Severity        Severity
		Type            Type
		Location        string
		Position        *Position
		Wrapped         *StackTrace
		Err             error
		Message         string
		WrappingMessage string
		Info            StructInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test simple",
			fields: fields{
				Severity: SeverityError,
				Type:     TypeValidating,
				Location: "/tmp/location.raml",
				Message:  "error message",
			},
			want: "[error] validating: /tmp/location.raml:1\n\terror message",
		},
		{
			name: "Test with wrapped",
			fields: fields{
				Severity:        SeverityError,
				Type:            TypeValidating,
				Location:        "/tmp/location.raml",
				Position:        &Position{1, 2},
				Message:         "error message",
				WrappingMessage: "wrapping message",
				Wrapped: &StackTrace{
					Severity: SeverityCritical,
					Type:     TypeParsing,
					Location: "/tmp/location2.raml",
					Position: &Position{3, 4},
					Message:  "error message 2",
				},
			},
			want: "[error] validating: /tmp/location.raml:1:2\n\twrapping message: error message\n[critical] parsing: /tmp/location2.raml:3:4\n\terror message 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &StackTrace{
				Severity:        tt.fields.Severity,
				Type:            tt.fields.Type,
				Location:        tt.fields.Location,
				Position:        tt.fields.Position,
				Wrapped:         tt.fields.Wrapped,
				Err:             tt.fields.Err,
				Message:         tt.fields.Message,
				WrappingMessage: tt.fields.WrappingMessage,
				Info:            tt.fields.Info,
			}
			if got := e.Sprint(); got != tt.want {
				t.Errorf("Sprint() = %v, want %v", got, tt.want)
			}
		})
	}
}
