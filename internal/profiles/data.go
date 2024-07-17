package profiles

import (
	"encoding/json"
	"io"
)

type Profile struct {
	Text     *string                `json:"text"`
	Metadata map[string]interface{} `json:"metadata"`
}

func DecodeProfile(r io.Reader) (Profile, error) {
	var profile Profile
	if err := json.NewDecoder(r).Decode(&profile); err != nil {
		return profile, err
	}
	return profile, nil
}

func ValidateProfile(profile Profile) []string {
	errMsgs := []string{}
	if profile.Text == nil {
		errMsgs = append(errMsgs, "The field \"text\" is required")
	}
	return errMsgs
}
