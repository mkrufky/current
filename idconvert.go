package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidExternalID is returned by internalID() when the provided string is an invalid external ID
var ErrInvalidExternalID = errors.New("invalid external ID")

// ErrIDNotFound is returned when the ID is not found in the datastore
var ErrIDNotFound = errors.New("id not found")

func internalID(s string) (int64, error) {
	a := strings.Split(s, "-")
	if len(a) == 0 {
		return -1, ErrInvalidExternalID
	}
	return strconv.ParseInt(a[len(a)-1], 10, 64)
}

func externalID(d int) visitID {
	return visitID{fmt.Sprintf("some-visit-id-%d", d)}
}
