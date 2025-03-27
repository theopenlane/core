> This project was originally inspired by, and parts taken from: https://github.com/adelowo/gulter
> Material additions / changes are in the form of support for use within GraphQL and REST APIs, functional options, and additional AWS S3 support features

# Objects

`Objects` is a package designed to simplify the process of managing receipt of
multipart/form-data requests and subsequently the uploading of files.

## Usage

Assuming you have a HTML form like this:

```html
<form action="/" method="post" enctype="multipart/form-data">
  <input type="file" name="form-field-1" />
  <input type="file" name="form-field-2" />
</form>
```

To create a new `objects` instance, you can do something like this:

```go
objectManager, _ := objects.New(
 objects.WithMaxFileSize(10<<20),
 objects.WithKeys([]string{"form-field-1", "form-field-2"}),
 objects.WithUploaderFunc(
   func(ctx context.Context, u *Objects, files []FileUpload) ([]File, error) {
    // add your own custom uploader functionality here
    // or leave out to use the default uploader func
   }
 ),
 objects.WithValidationFunc(
  objects.ChainValidators(objects.MimeTypeValidator("image/jpeg", "image/png"),
   func(f objects.File) error {
    // Your own custom validation function on the file here
    return nil
   })),
 objects.WithStorage(s3Store),
)
```

The `objectManager` can be used with the provided middleware
`FileUploadMiddleware(objectManager)` and added to the chain of middleware
within the server. This is a generic `http.Handler` so it can be used with your
router of choice. For example, to be used with `echo`:

```go
echo.WrapMiddleware(objects.FileUploadMiddleware(objectManager))
```

### Standard HTTP router

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCreds "github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func main() {
	s3Store, err := storage.NewS3FromConfig(aws.Config{
		Region: "us-west-2",
		Credentials: awsCreds.NewStaticCredentialsProvider(
			"accessKey", "secretKey", ""),
	}, storage.S3Options{
		Bucket: "std-router",
	})
	if err != nil {
		panic(err.Error())
	}

	objectsHandler, err := objects.New(
		objects.WithMaxFileSize(10<<20),
		objects.WithStorage(s3Store),
		objects.WithKeys([]string{"name", "mitb"}),
	)

	mux := http.NewServeMux()

	// upload all files with the "name" and "mitb" fields on this route
	mux.Handle("/", objects.FileUploadMiddleware(objectsHandler)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Uploading file")

		ctx := r.Context()

		// return all uploaded files
		f, err := objects.FilesFromContext(ctx)
		if err != nil {
			fmt.Println(err)
			return
		}

		// return uploaded files with the form field "mitb"
		ff, err := objects.FilesFromContextWithKey(ctx, "mitb")
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("%+v", ff)

		for _, v := range f {
			fmt.Printf("%+v", v)
			fmt.Println()
		}
	})))

	http.ListenAndServe(":3300", mux)
}
```

### Ignoring non existent keys in the multipart Request

Sometimes, the keys you have configured the middleware might get dropped from
the frontend for some reason, ideally the middleware fails if it cannot find a
configured key in the request. To disable this behavior and ignore the missing
key, you can make use of the `WithIgnoreNonExistentKey(true)` option to prevent
the middleware from causing an error when such keys do not exists

### Customizing the error response

Since `Objects` could be used as a middleware, it returns an error to the client
if found, this might not match your existing structure, so to configure the
response, use the `WithErrorResponseHandler`. The default is shown below and can
be used as a template to define yours.

```go
var errHandler objects.ErrResponseHandler = func(err error, statusCode int) http.HandlerFunc {
  return func(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    fmt.Fprintf(w, `{"message" : "could not upload file", "error" : "%s"}`, err.Error())
  }
}

// add the following to the objects options:
// objects.WithErrorResponseHandler(errHandler)
```

### Writing your own custom upload logic

The uploader function by default will just upload the file to the storage
backend. In some cases you may want custom logic, e.g. update your local
database with information. A custom `UploaderFunc` can be used to do this.

```go
var uploader objects.UploaderFunc = func(ctx context.Context, u *objects.Objects, files []objects.FileUpload) ([]objects.File, error) {
	uploadedFiles := make([]objects.File, 0, len(files))

	for _, f := range files {
		// do things
	}

	return uploadedFiles, nil
}

// add the following to the objects options:
// objects.WithUploaderFunc(uploader)
```

### Writing your custom validator logic

Sometimes, you could have some custom logic to validate uploads, in this example
below, we limit the size of the upload based on the mimeypes of the uploaded
files

```go
var customValidator objects.ValidationFunc = func(f objects.File) error {
 switch f.MimeType {
 case "image/png":
  if f.Size > 4096 {
   return errors.New("file size too large")
  }

  return nil

 case "application/pdf":
  if f.Size > (1024 * 10) {
   return errors.New("file size too large")
  }

  return nil
 default:
  return nil
 }
}

// add the following to the objects options:
// objects.WithValidationFunc(customValidator)
```
