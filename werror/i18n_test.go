package werror

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func TestNewI18nErrTmpl(t *testing.T) {
	tests := []struct {
		name       string
		base       Err
		i18n       *i18n.Message
		wantErr    bool
		wantCode   string
		wantStatus int
	}{
		{
			name: "success with i18n ID and template variables",
			base: ErrBadRequest,
			i18n: &i18n.Message{
				ID:    "UserNotFound",
				Other: "User {{.Name}} not found",
			},
			wantErr:    false,
			wantCode:   "UserNotFound",
			wantStatus: 400,
		},
		{
			name: "success with empty i18n ID - static message",
			base: ErrBadRequest,
			i18n: &i18n.Message{
				ID:    "",
				Other: "Bad argument",
			},
			wantErr:    false,
			wantCode:   "",
			wantStatus: 400,
		},
		{
			name: "success with whitespace i18n ID - static message",
			base: ErrBadRequest,
			i18n: &i18n.Message{
				ID:    "   ",
				Other: "Bad argument",
			},
			wantErr:    false,
			wantCode:   "   ",
			wantStatus: 400,
		},
		{
			name: "error - i18n.Other is empty",
			base: ErrBadRequest,
			i18n: &i18n.Message{
				ID:    "TestError",
				Other: "",
			},
			wantErr: true,
		},
		{
			name: "success - whitespace Other is valid (static message)",
			base: ErrBadRequest,
			i18n: &i18n.Message{
				ID:    "TestError",
				Other: "   ",
			},
			wantErr:    false,
			wantCode:   "TestError",
			wantStatus: 400,
		},
		{
			name: "error - invalid template syntax",
			base: ErrBadRequest,
			i18n: &i18n.Message{
				ID:    "TestError",
				Other: "User {{.Name not found",
			},
			wantErr: true,
		},
		{
			name: "success - different base errors with template variables",
			base: ErrNotFound,
			i18n: &i18n.Message{
				ID:    "ResourceNotFound",
				Other: "Resource {{.ID}} not found",
			},
			wantErr:    false,
			wantCode:   "ResourceNotFound",
			wantStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewI18nErrTmpl(tt.base, tt.i18n)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewI18nErrTmpl() expected error, got nil")
				}
				if got != nil {
					t.Errorf("NewI18nErrTmpl() should return nil I18nErrTmpl on error, got %v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewI18nErrTmpl() unexpected error = %v", err)
			}
			if got == nil {
				t.Fatal("NewI18nErrTmpl() returned nil I18nErrTmpl")
			}

			if got.GetI18n().ID != tt.wantCode {
				t.Errorf("I18nErrTmpl.GetI18n().ID = %v, want %v", got.GetI18n().ID, tt.wantCode)
			}

			if got.GetBase().GetHttpStatus() != tt.wantStatus {
				t.Errorf("I18nErrTmpl.GetBase().GetHttpStatus() = %v, want %v",
					got.GetBase().GetHttpStatus(), tt.wantStatus)
			}

			if got.GetTemplate() == nil {
				t.Error("I18nErrTmpl.GetTemplate() returned nil")
			}
		})
	}
}

func TestNewI18nErrTmpl_ErrI18nMessageOtherMissing(t *testing.T) {
	base := ErrBadRequest
	i18nMsg := &i18n.Message{
		ID:    "TestError",
		Other: "",
	}

	_, err := NewI18nErrTmpl(base, i18nMsg)

	if !errors.Is(err, ErrI18nMessageOtherMissing) {
		t.Errorf("NewI18nErrTmpl() error = %v, want %v", err, ErrI18nMessageOtherMissing)
	}
}

func TestMustNewI18nErrTmpl(t *testing.T) {
	t.Run("success - valid i18n message", func(t *testing.T) {
		base := ErrBadRequest
		i18nMsg := &i18n.Message{
			ID:    "UserNotFound",
			Other: "User {{.Name}} not found",
		}

		got := MustNewI18nErrTmpl(base, i18nMsg)

		if got == nil {
			t.Fatal("MustNewI18nErrTmpl() returned nil I18nErrTmpl")
		}

		if got.GetI18n().ID != "UserNotFound" {
			t.Errorf("I18nErrTmpl.GetI18n().ID = %v, want %v", got.GetI18n().ID, "UserNotFound")
		}
	})

	t.Run("panic - i18n.Other is missing", func(t *testing.T) {
		base := ErrBadRequest
		i18nMsg := &i18n.Message{
			ID:    "TestError",
			Other: "",
		}

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNewI18nErrTmpl() should panic with missing Other, but did not")
			}
		}()

		MustNewI18nErrTmpl(base, i18nMsg)
	})

	t.Run("panic - invalid template syntax", func(t *testing.T) {
		base := ErrBadRequest
		i18nMsg := &i18n.Message{
			ID:    "TestError",
			Other: "User {{.Name not found",
		}

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNewI18nErrTmpl() should panic with invalid template, but did not")
			}
		}()

		MustNewI18nErrTmpl(base, i18nMsg)
	})
}

func TestI18nErrTmpl_Render(t *testing.T) {
	tests := []struct {
		name         string
		i18nMsg      *i18n.Message
		templateData any
		wantMsg      string
		wantCode     string
		wantErr      bool
	}{
		{
			name: "success - render with string data",
			i18nMsg: &i18n.Message{
				ID:    "UserNotFound",
				Other: "User {{.Name}} not found",
			},
			templateData: map[string]string{
				"Name": "Alice",
			},
			wantMsg:  "User Alice not found",
			wantCode: "UserNotFound",
			wantErr:  false,
		},
		{
			name: "success - render with multiple variables",
			i18nMsg: &i18n.Message{
				ID:    "ResourceNotFound",
				Other: "{{.ResourceType}} {{.ID}} not found",
			},
			templateData: map[string]string{
				"ResourceType": "File",
				"ID":           "123",
			},
			wantMsg:  "File 123 not found",
			wantCode: "ResourceNotFound",
			wantErr:  false,
		},
		{
			name: "success - render without variables",
			i18nMsg: &i18n.Message{
				ID:    "SimpleError",
				Other: "A simple error occurred",
			},
			templateData: nil,
			wantMsg:      "A simple error occurred",
			wantCode:     "SimpleError",
			wantErr:      false,
		},
		{
			name: "success - missing variable outputs <no value>",
			i18nMsg: &i18n.Message{
				ID:    "UserNotFound",
				Other: "User {{.Name}} not found",
			},
			templateData: map[string]string{
				"OtherField": "value",
			},
			wantMsg:  "User <no value> not found",
			wantCode: "UserNotFound",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := NewI18nErrTmpl(ErrBadRequest, tt.i18nMsg)
			if err != nil {
				t.Fatalf("NewI18nErrTmpl() failed: %v", err)
			}

			gotErr, err := tmpl.Render(tt.templateData)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Render() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Render() unexpected error = %v", err)
			}

			if gotErr == nil {
				t.Fatal("Render() returned nil I18nErr")
			}

			if gotErr.GetMessage() != tt.wantMsg {
				t.Errorf("I18nErr.GetMessage() = %v, want %v", gotErr.GetMessage(), tt.wantMsg)
			}

			if gotErr.GetCode() != tt.wantCode {
				t.Errorf("I18nErr.GetCode() = %v, want %v", gotErr.GetCode(), tt.wantCode)
			}

			// Verify GetRenderedData returns the data used for rendering
			if !reflect.DeepEqual(gotErr.GetRenderedData(), tt.templateData) {
				t.Errorf("I18nErr.GetRenderedData() = %v, want %v",
					gotErr.GetRenderedData(), tt.templateData)
			}
		})
	}
}

func TestI18nErrTmpl_RenderMultipleTimes(t *testing.T) {
	i18nMsg := &i18n.Message{
		ID:    "UserNotFound",
		Other: "User {{.Name}} not found",
	}

	tmpl, err := NewI18nErrTmpl(ErrBadRequest, i18nMsg)
	if err != nil {
		t.Fatalf("NewI18nErrTmpl() failed: %v", err)
	}

	// Render first time
	err1, err := tmpl.Render(map[string]string{"Name": "Alice"})
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	// Render second time
	err2, err := tmpl.Render(map[string]string{"Name": "Bob"})
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	// Each error should be independent
	if err1.GetMessage() != "User Alice not found" {
		t.Errorf("First error message = %v, want 'User Alice not found'", err1.GetMessage())
	}

	if err2.GetMessage() != "User Bob not found" {
		t.Errorf("Second error message = %v, want 'User Bob not found'", err2.GetMessage())
	}

	// Verify rendered data is different for each
	data1 := err1.GetRenderedData().(map[string]string)
	data2 := err2.GetRenderedData().(map[string]string)

	if data1["Name"] != "Alice" {
		t.Errorf("First error data Name = %v, want 'Alice'", data1["Name"])
	}

	if data2["Name"] != "Bob" {
		t.Errorf("Second error data Name = %v, want 'Bob'", data2["Name"])
	}
}

func TestNewI18nErr(t *testing.T) {
	// NewI18nErr is a convenience function that creates template and renders with nil data
	tests := []struct {
		name       string
		base       Err
		i18n       *i18n.Message
		wantErr    bool
		wantCode   string
		wantMsg    string
		wantStatus int
	}{
		{
			name: "success with variables - renders with nil data",
			base: ErrBadRequest,
			i18n: &i18n.Message{
				ID:    "UserNotFound",
				Other: "User {{.Name}} not found",
			},
			wantErr:    false,
			wantCode:   "UserNotFound",
			wantMsg:    "User <no value> not found",
			wantStatus: 400,
		},
		{
			name: "success without variables",
			base: ErrBadRequest,
			i18n: &i18n.Message{
				ID:    "SimpleError",
				Other: "A simple error occurred",
			},
			wantErr:    false,
			wantCode:   "SimpleError",
			wantMsg:    "A simple error occurred",
			wantStatus: 400,
		},
		{
			name: "error - i18n.Other is empty",
			base: ErrBadRequest,
			i18n: &i18n.Message{
				ID:    "TestError",
				Other: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewI18nErr(tt.base, tt.i18n, nil)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewI18nErr() expected error, got nil")
				}
				if got != nil {
					t.Errorf("NewI18nErr() should return nil I18nErr on error, got %v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewI18nErr() unexpected error = %v", err)
			}
			if got == nil {
				t.Fatal("NewI18nErr() returned nil I18nErr")
			}

			if got.GetCode() != tt.wantCode {
				t.Errorf("I18nErr.GetCode() = %v, want %v", got.GetCode(), tt.wantCode)
			}

			if got.GetMessage() != tt.wantMsg {
				t.Errorf("I18nErr.GetMessage() = %v, want %v", got.GetMessage(), tt.wantMsg)
			}

			if got.GetHttpStatus() != tt.wantStatus {
				t.Errorf("I18nErr.GetHttpStatus() = %v, want %v", got.GetHttpStatus(), tt.wantStatus)
			}
		})
	}
}

func TestMustNewI18nErr(t *testing.T) {
	t.Run("success - valid i18n message", func(t *testing.T) {
		base := ErrBadRequest
		i18nMsg := &i18n.Message{
			ID:    "UserNotFound",
			Other: "User {{.Name}} not found",
		}

		got := MustNewI18nErr(base, i18nMsg, nil)

		if got == nil {
			t.Fatal("MustNewI18nErr() returned nil I18nErr")
		}

		if got.GetCode() != "UserNotFound" {
			t.Errorf("I18nErr.GetCode() = %v, want %v", got.GetCode(), "UserNotFound")
		}
	})

	t.Run("panic - i18n.Other is missing", func(t *testing.T) {
		base := ErrBadRequest
		i18nMsg := &i18n.Message{
			ID:    "TestError",
			Other: "",
		}

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNewI18nErr() should panic with missing Other, but did not")
			}
		}()

		MustNewI18nErr(base, i18nMsg, nil)
	})
}

func TestSi18nerr_GetI18n(t *testing.T) {
	i18nMsg := &i18n.Message{
		ID:    "TestError",
		Other: "Test error message",
	}

	si18nerr := &Si18nerr{
		Serr: Serr{},
		i18n: i18nMsg,
	}

	got := si18nerr.GetI18n()

	if got != i18nMsg {
		t.Errorf("GetI18n() = %v, want %v", got, i18nMsg)
	}
}

func TestSi18nerr_GetRenderedData(t *testing.T) {
	testData := map[string]string{"Name": "Alice"}

	si18nerr := &Si18nerr{
		Serr:         Serr{},
		i18n:         &i18n.Message{},
		renderedData: testData,
	}

	got := si18nerr.GetRenderedData()

	if !reflect.DeepEqual(got, testData) {
		t.Errorf("GetRenderedData() = %v, want %v", got, testData)
	}
}

func TestSi18nerr_ImplementsI18nErr(t *testing.T) {
	// Test that *Si18nerr implements I18nErr interface
	var _ I18nErr = &Si18nerr{}
}

func TestI18nErr_ErrInterface(t *testing.T) {
	// Test that I18nErr implements Err interface
	i18nMsg := &i18n.Message{
		ID:    "TestError",
		Other: "Test error message",
	}

	i18nErr, err := NewI18nErr(ErrBadRequest, i18nMsg, nil)
	if err != nil {
		t.Fatalf("NewI18nErr() failed: %v", err)
	}

	// Test Err interface methods
	var errInterface Err = i18nErr

	if errInterface.GetCode() != "TestError" {
		t.Errorf("GetCode() = %v, want 'TestError'", errInterface.GetCode())
	}

	if errInterface.GetHttpStatus() != 400 {
		t.Errorf("GetHttpStatus() = %v, want 400", errInterface.GetHttpStatus())
	}

	// Test SetCode
	errInterface.SetCode("NewCode")
	if errInterface.GetCode() != "NewCode" {
		t.Errorf("SetCode() failed, got %v, want 'NewCode'", errInterface.GetCode())
	}

	// Test SetMessage
	errInterface.SetMessage("New message")
	if errInterface.GetMessage() != "New message" {
		t.Errorf("SetMessage() failed, got %v, want 'New message'", errInterface.GetMessage())
	}
}

func TestI18nErr_ErrorContainsRenderedMessage(t *testing.T) {
	// Test that the Error() output contains the rendered message
	i18nMsg := &i18n.Message{
		ID:    "UserNotFound",
		Other: "User {{.Name}} not found",
	}

	tmpl, err := NewI18nErrTmpl(ErrBadRequest, i18nMsg)
	if err != nil {
		t.Fatalf("NewI18nErrTmpl() failed: %v", err)
	}

	i18nErr, err := tmpl.Render(map[string]string{"Name": "Bob"})
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	errorOutput := i18nErr.Error()

	if !strings.Contains(errorOutput, "User Bob not found") {
		t.Errorf("Error() should contain rendered message, got: %v", errorOutput)
	}

	if !strings.Contains(errorOutput, "400") {
		t.Errorf("Error() should contain HTTP status, got: %v", errorOutput)
	}
}

func TestI18nErr_MetadataContainsTemplateData(t *testing.T) {
	// Test that Metadata is set to the template data
	i18nMsg := &i18n.Message{
		ID:    "UserNotFound",
		Other: "User {{.Name}} not found",
	}

	tmpl, err := NewI18nErrTmpl(ErrBadRequest, i18nMsg)
	if err != nil {
		t.Fatalf("NewI18nErrTmpl() failed: %v", err)
	}

	data := map[string]string{"Name": "Alice"}
	i18nErr, err := tmpl.Render(data)
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	if !reflect.DeepEqual(i18nErr.GetMetadata(), data) {
		t.Errorf("GetMetadata() = %v, want %v", i18nErr.GetMetadata(), data)
	}
}
