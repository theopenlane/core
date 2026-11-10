// Package relocate provides the one-time backfill that moves file records still pointing at S3
// onto the R2 provider and purges the records that have no reason to be moved
package relocate
