package handlers

import (
	"bytes"
	"context"
	"errors"
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

type fakeErroringJobRepo struct{}

func (r fakeErroringJobRepo) Create(ctx context.Context, profile schemav1.Profile) (string, error) {
	return "", errors.New("Job backend go boom!")
}

func TestServeHTTP(t *testing.T) {
	handler := ProfilesHandler{fakeJobRepo{}}
	body := []byte(`{"text": "foo", "images": ["a", "b"], "metadata": {"id": 1}}`)
	req, _ := http.NewRequest("POST", "/api/v1/profiles", bytes.NewBuffer(body))
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusCreated)
	expected := fmt.Sprint("{\"id\":\"", TestJobName, "\"}\n")
	assert.Equal(t, resp.Body.String(), expected)
}

func TestServeHTTPWithBadJSON(t *testing.T) {
	handler := ProfilesHandler{fakeJobRepo{}}
	body := []byte(`{text: "foo"}`)
	req, _ := http.NewRequest("POST", "/api/v1/profiles", bytes.NewBuffer(body))
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusBadRequest)
	expected := "Deserialization error: invalid character 't' looking for beginning of object key string\n"
	assert.Equal(t, resp.Body.String(), expected)
}

func TestServeHTTPWithBadSchema(t *testing.T) {
	handler := ProfilesHandler{fakeJobRepo{}}
	body := []byte(`{"body": "foo", "images": ["a", "b"], "metadata": {"id": 1}}`)
	req, _ := http.NewRequest("POST", "/api/v1/profiles", bytes.NewBuffer(body))
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusBadRequest)
	expected := "The field \"text\" is required\n"
	assert.Equal(t, resp.Body.String(), expected)
}

func TestServeHTTPWithErrorInJobRepo(t *testing.T) {
	handler := ProfilesHandler{fakeErroringJobRepo{}}
	body := []byte(`{"text": "foo", "images": ["a", "b"], "metadata": {"id": 1}}`)
	req, _ := http.NewRequest("POST", "/api/v1/profiles", bytes.NewBuffer(body))
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	expected := "Error creating Kubernetes Job\n"
	assert.Equal(t, resp.Body.String(), expected)
}
