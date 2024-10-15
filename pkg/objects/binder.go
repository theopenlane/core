package objects

import (
	"fmt"
	"mime/multipart"
	"reflect"
	"strings"

	echo "github.com/theopenlane/echox"
)

var (
	typeMultipartFileHeader      = reflect.TypeOf((*multipart.FileHeader)(nil))
	typeMultipartSliceFileHeader = reflect.TypeOf(([]*multipart.FileHeader)(nil))
)

// BindFunc is a function that binds the request body to the struct pointer
type BindFunc func(echo.Context, interface{}) error

// Bind binds the request body to the struct pointer
func (fn BindFunc) Bind(ctx echo.Context, i interface{}) error {
	return fn(ctx, i)
}

// NewFileBinder returns a new FileBinder that binds the request body to the struct pointer and bind the form files to the struct fields.
func NewFileBinder(b echo.Binder) echo.Binder {
	return BindFunc(func(ctx echo.Context, i interface{}) error {
		if err := b.Bind(ctx, i); err != nil {
			return err
		}

		ctype := ctx.Request().Header.Get(echo.HeaderContentType)

		// if bind form
		if strings.HasPrefix(ctype, echo.MIMEApplicationForm) || strings.HasPrefix(ctype, echo.MIMEMultipartForm) {
			// get form files
			var form *multipart.Form

			form, err := ctx.MultipartForm()
			if err != nil {
				return err
			}

			return BindFile(i, ctx, form.File)
		}

		return nil
	})
}

// BindFile binds the form files to the struct fields
func BindFile(i interface{}, ctx echo.Context, files map[string][]*multipart.FileHeader) error {
	iValue := reflect.Indirect(reflect.ValueOf(i))
	// check bind type is struct pointer
	if iValue.Kind() != reflect.Struct {
		return fmt.Errorf("%w: BindFile input not is struct pointer, indirect type is %s", ErrUnexpectedType, iValue.Type().String())
	}

	iType := iValue.Type()
	for i := range iType.NumField() {
		fType := iType.Field(i)

		// check canset field
		fValue := iValue.Field(i)
		if !fValue.CanSet() {
			continue
		}
		// revc type must *multipart.FileHeader or []*multipart.FileHeader
		switch fType.Type {
		case typeMultipartFileHeader:
			file := getFiles(files, fType.Name, fType.Tag.Get("form"))
			if len(file) > 0 {
				fValue.Set(reflect.ValueOf(file[0]))
			}
		case typeMultipartSliceFileHeader:
			file := getFiles(files, fType.Name, fType.Tag.Get("form"))
			if len(file) > 0 {
				fValue.Set(reflect.ValueOf(file))
			}
		}
	}

	return nil
}

// getFiles returns the files by the field name or form name
func getFiles(files map[string][]*multipart.FileHeader, names ...string) []*multipart.FileHeader {
	for _, name := range names {
		file, ok := files[name]
		if ok {
			return file
		}
	}

	return nil
}
