package werror

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	ErrI18nMessageOtherMissing = errors.New("i18n.Message.Other is missing")
	ErrI18nTemplateMissing     = errors.New("i18nTmpl is missing")
)

// I18nErrTmpl is an i18n template that can render multiple I18nErr instances.
type I18nErrTmpl struct {
	base Err
	i18n *i18n.Message
	tmpl *template.Template
}

// I18nErr is the error interface with i18n support.
// It represents a rendered error with an immutable message.
type I18nErr interface {
	Err
	GetI18n() *i18n.Message
	GetRenderedData() any
}

// Si18nerr is the concrete implementation of I18nErr (rendered error).
//
//nolint:errname // ignore
type Si18nerr struct {
	Serr

	i18n         *i18n.Message
	renderedData any
}

// NewI18nErrTmpl creates an I18nErrTmpl from i18n.Message.
// The i18n.ID will be used as the error code.
// The i18n.Other will be parsed as a Go template.
func NewI18nErrTmpl(base Err, i18n *i18n.Message) (*I18nErrTmpl, error) {
	if i18n.Other == "" {
		return nil, ErrI18nMessageOtherMissing
	}

	tmpl, err := template.New(i18n.ID).Parse(i18n.Other)
	if err != nil {
		return nil, err
	}

	return &I18nErrTmpl{
		base: base,
		i18n: i18n,
		tmpl: tmpl,
	}, nil
}

// MustNewI18nErrTmpl creates an I18nErrTmpl and panics on error.
func MustNewI18nErrTmpl(base Err, i18n *i18n.Message) *I18nErrTmpl {
	tmpl, err := NewI18nErrTmpl(base, i18n)
	if err != nil {
		panic(err)
	}
	return tmpl
}

// Render creates a new I18nErr with the template executed using templateData.
func (t *I18nErrTmpl) Render(templateData any) (I18nErr, error) {
	if t.tmpl == nil {
		return nil, ErrI18nTemplateMissing
	}

	var buf bytes.Buffer
	if err := t.tmpl.Execute(&buf, templateData); err != nil {
		return nil, err
	}
	msg := buf.String()

	// Create the rendered error
	err := NewErr(t.base, msg, "")
	// Use i18n ID as code
	if strings.TrimSpace(t.i18n.ID) != "" {
		err.SetCode(t.i18n.ID)
	}
	ierr := &Si18nerr{
		//nolint:errcheck // type must match
		Serr:         *err.(*Serr),
		i18n:         t.i18n,
		renderedData: templateData,
	}
	ierr.SetMetadata(templateData)

	return ierr, nil
}

func (t *I18nErrTmpl) GetI18n() *i18n.Message {
	return t.i18n
}

func (t *I18nErrTmpl) GetTemplate() *template.Template {
	return t.tmpl
}

func (t *I18nErrTmpl) GetBase() Err {
	return t.base
}

// GetI18n returns the original i18n message.
func (e *Si18nerr) GetI18n() *i18n.Message {
	return e.i18n
}

// GetRenderedData returns the data used to render the error message.
func (e *Si18nerr) GetRenderedData() any {
	return e.renderedData
}

// NewI18nErr creates a rendered I18nErr from i18n.Message.
// For simple messages without template variables (no "{{"), creates the error directly.
// templateData is used only if the message contains template variables.
// This is a convenience function for simple cases.
func NewI18nErr(base Err, i18n *i18n.Message, templateData any) (I18nErr, error) {
	if i18n.Other == "" {
		return nil, ErrI18nMessageOtherMissing
	}

	// Fast path: no template variables, create error directly
	if !strings.Contains(i18n.Other, "{{") {
		code := i18n.ID
		if strings.TrimSpace(code) == "" {
			code = base.GetCode()
		}
		return &Si18nerr{
			Serr: Serr{
				error:      fmt.Errorf("%w: %s", base, i18n.Other),
				HttpStatus: base.GetHttpStatus(),
				Code:       code,
				Message:    i18n.Other,
			},
			i18n:         i18n,
			renderedData: templateData,
		}, nil
	}

	// Slow path: has template variables, use template
	tmpl, err := NewI18nErrTmpl(base, i18n)
	if err != nil {
		return nil, err
	}
	return tmpl.Render(templateData)
}

// MustNewI18nErr creates a rendered I18nErr and panics on error.
func MustNewI18nErr(base Err, i18n *i18n.Message, templateData any) I18nErr {
	err, e := NewI18nErr(base, i18n, templateData)
	if e != nil {
		panic(e)
	}
	return err
}
