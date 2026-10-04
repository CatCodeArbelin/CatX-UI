//go:build !cgo

package sponsors

func isSQLiteSponsorDuplicateError(error) bool { return false }
