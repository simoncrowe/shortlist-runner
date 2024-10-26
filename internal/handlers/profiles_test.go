package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	schemav1 "github.com/simoncrowe/shortlist-schema/lib/v1"
	"github.com/stretchr/testify/assert"
)

const TestJobName = "assessor-cfe8eb87-b9d1-4a84-a253-510228a6241c"

type fakeJobRepo struct{}

func (r fakeJobRepo) Create(ctx context.Context, profile schemav1.Profile) (string, error) {
	return TestJobName, nil
}

func TestServeHTTP(t *testing.T) {
	handler := ProfilesHandler{fakeJobRepo{}}
	body := []byte(`{"text": "foo", "images": ["a", "b"], "metadata": {"id": 1}}`)
	req, _ := http.NewRequest("POST", "/api/v1/profiles", bytes.NewBuffer(body))
	respRec := httptest.NewRecorder()

	handler.ServeHTTP(respRec, req)

	assert.Equal(t, respRec.Code, http.StatusCreated)
	expected := fmt.Sprint("{\"id\":\"", TestJobName, "\"}\n")
	assert.Equal(t, respRec.Body.String(), expected)
}
