# Objects

`Objects` is a package designed to simplify the process of managing receipt of multipart/form-data requests and subsequently the uploading of files.


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
 handler, _ := objects.New(
  objects.WithMaxFileSize(10<<20),
  objects.WithValidationFunc(
   objects.ChainValidators(objects.MimeTypeValidator("image/jpeg", "image/png"),
    func(f objects.File) error {
     // Your own custom validation function on the file here
     return nil
    })),
  objects.WithStorage(s3Store),
 )
```

The `handler` is really just a HTTP middleware with the following signature
`Upload(keys ...string) func(next http.Handler) http.Handler`. `keys` here
are the input names from the HTML form, so you can chain this into almost any HTTP
router.

### Standard HTTP router

```go
package main

import (
 "fmt"
 "net/http"

 "github.com/theopenlane/core/pkg/objects"
 "github.com/theopenlane/core/pkg/objects/storage"
)

func main() {
 s3Store, err := storage.NewS3FromEnvironment(storage.S3Options{
  Bucket: "std-router",
 })
 if err != nil {
  panic(err.Error())
 }

 handler, err := objects.New(
  objects.WithMaxFileSize(10<<20),
  objects.WithStorage(s3Store),
 )

 mux := http.NewServeMux()

 // upload all files with the "name" and "mitb" fields on this route
 mux.Handle("/", handler.Upload("name", "mitb")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  fmt.Println("Uploaded file")

  // return all uploaded files
  f, err := objects.FilesFromContext(r)
  if err != nil {
   fmt.Println(err)
   return
  }

  // return uploaded files with the form field "mitb"
  ff, err := objects.FilesFromContextWithKey(r, "mitb")
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

Sometimes, the keys you have configured the middleware might get dropped from the
frontend for some reason, ideally the middleware fails if it cannot find a
configured key in the request. To disable this behavior and ignore the missing
key, you can make use of the `WithIgnoreNonExistentKey(true)` option to prevent the
middleware from causing an error when such keys do not exists

### Customizing the error response

Since `Objects` is could be used as a middleware, it returns an error to the client if found,
this might not match your existing structure, so to configure the response, use the
`WithErrorResponseHandler`. The default is shown below and can be used as a template
to define yours.

```go

 errHandler ErrResponseHandler = func(err error) http.HandlerFunc {
  return func(w http.ResponseWriter, _ *http.Request) {
   w.Header().Set("Content-Type", "application/json")
   w.WriteHeader(http.StatusInternalServerError)
   fmt.Fprintf(w, `{"message" : "could not upload file", "error" : %s}`, err.Error())
  }
 }
```

### Writing your custom validator logic

Sometimes, you could have some custom logic to validate uploads, in this example
below, we limit the size of the upload based on the mimeypes of the uploaded files

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

```
