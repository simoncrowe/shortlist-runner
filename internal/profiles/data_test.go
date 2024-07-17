package profiles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingText(t *testing.T) {
	profile := Profile{Text: nil}

	errMsgs := ValidateProfile(profile)

	expectedErrMsgs := []string{"The field \"text\" is required"}
	assert.Equal(t, errMsgs, expectedErrMsgs)
}

func TestValidProfile(t *testing.T) {
	text := "foo"
	profile := Profile{Text: &text}

	errMsgs := ValidateProfile(profile)

	expectedErrMsgs := []string{}
	assert.Equal(t, errMsgs, expectedErrMsgs)
}
