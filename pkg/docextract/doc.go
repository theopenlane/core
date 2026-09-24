// Package docextract extracts structured data from uploaded documents with a model.
//
// The package is generic over the kind of document: a Kind supplies the prompt and response
// schema for each section it can extract, and optionally a Streamer for sections too large for
// one response, which merges streamed payloads and asks the model to continue where it stopped.
// A Client uploads the document once per extraction and runs the request loop.
//
// Validation runs before any model call. A Profile registered for a kind checks the page count
// and a weighted table of markers against the document's text layer, so an upload that is not
// the expected kind of document fails fast with a reason safe to show to the uploader.
//
// Document kinds live in subpackages such as soc2, which register their profile on init and
// carry their own prompt configuration
package docextract
