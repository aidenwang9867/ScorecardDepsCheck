package main

import "errors"

// static Errors for mapping
var (
	errMappingNotFound = errors.New("ecosystem mapping not found")
	errInvalid         = errors.New("invalid")
)
