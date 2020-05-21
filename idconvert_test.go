package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInternalID(t *testing.T) {

	a := []struct {
		s string
		r int
		e error
	}{
		{
			"some-visit-id-3",
			3,
			nil,
		},
		{
			"some-visit-id-0",
			0,
			nil,
		},
		{
			"some-visit-id-02",
			2,
			nil,
		},
		{
			"some-visit-id-002",
			2,
			nil,
		},
		{
			"some-visit-id-000",
			0,
			nil,
		},
		{
			"some-visit-id-0020",
			20,
			nil,
		},
		{
			"crazy-visit-number-30",
			30,
			nil,
		},
		{
			"50",
			50,
			nil,
		},
		{
			"some-other-visit-id-0020",
			20,
			nil,
		},
		{
			"visit-id-1",
			1,
			nil,
		},
		{
			"some-visit-id-missing",
			0,
			ErrInvalidExternalID,
		},
		{
			"missing",
			0,
			ErrInvalidExternalID,
		},
		{
			"missing",
			0,
			ErrInvalidExternalID,
		},
		{
			"O",
			0,
			ErrInvalidExternalID,
		},
		{
			"",
			0,
			ErrInvalidExternalID,
		},
	}

	for i, e := range a {
		v, err := internalID(e.s)
		if e.e == nil {
			assert.Nil(t, err, "%d) internalID(%s) should return nil error status", i, e.s)
			assert.Equal(t, int64(e.r), v, "%d) internalID(%s) should equal %d", i, e.s, e.r)
		} else {
			assert.NotNil(t, err, "%d) internalID(%s) should return non-nil error status", i, e.s)
		}
	}
}

func TestExternalID(t *testing.T) {

	a := []struct {
		i int
		r string
	}{
		{
			3,
			"some-visit-id-3",
		},
		{
			0,
			"some-visit-id-0",
		},
		{
			20,
			"some-visit-id-20",
		},
	}

	for i, e := range a {
		assert.Equal(t, visitID{e.r}, externalID(e.i), "%d) externalID(%d) should equal %s", i, e.i, e.r)
	}
}
