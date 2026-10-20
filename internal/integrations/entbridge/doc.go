// Package entbridge converts between ent records and the ent-free integration types, and gives
// first-party definitions access to the ent client behind the runtime services. Definitions that
// can be built outside core never import this package; only built-in definitions and the runtime do
package entbridge
