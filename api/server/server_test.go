package server_test

import (
	"crswty.com/cms/datastore"
	"crswty.com/cms/server"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var BasicType = server.Type{Name: "user", Id: "id", Schema:
// language=json
`{
	"$id": "http://example.com/schema/my-test-schema",
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"type": "object",
	"required": ["id", "name"],
	"properties": {
		"id": {
			"type": "string"
		},
		"name": {
			"type": "string"
		}
	}
}`}

func TestServer_List(t *testing.T) {

	store, err := datastore.NewMemory()
	require.NoError(t, err)
	require.NoError(t, store.Create(BasicType, "a", server.Object{"id": "a", "name": "123"}))
	require.NoError(t, store.Create(BasicType, "b", server.Object{"id": "b", "name": "456"}))

	url, closeFn := startServer(store, BasicType)
	defer closeFn()

	resp, err := http.Get(fmt.Sprintf("%s/%s", url, BasicType.Name))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	assert.Equal(t, resp.Header.Get("Content-Type"), "application/json")
	assert.JSONEq(t, `[{"id": "a", "name": "123"},{"id": "b", "name": "456"}]`, string(body))
}

func TestServer_Get(t *testing.T) {
	store, err := datastore.NewMemory()
	require.NoError(t, err)
	require.NoError(t, store.Create(BasicType, "1", server.Object{"id": 1, "name": "value1"}))
	require.NoError(t, store.Create(BasicType, "2", server.Object{"id": 2, "name": "value2"}))
	require.NoError(t, store.Create(BasicType, "50", server.Object{"id": 50, "name": "value50"}))

	url, closeFn := startServer(store, BasicType)
	defer closeFn()

	resp, err := http.Get(fmt.Sprintf("%s/%s/%s", url, BasicType.Name, "2"))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	assert.Equal(t, resp.Header.Get("Content-Type"), "application/json")
	assert.JSONEq(t, `{"id": 2, "name": "value2"}`, string(body))
}

func TestServer_Post(t *testing.T) {
	store, err := datastore.NewMemory()
	require.NoError(t, err)

	url, closeFn := startServer(store, BasicType)
	defer closeFn()

	resp, err := http.Post(fmt.Sprintf("%s/%s", url, BasicType.Name), "application/json", strings.NewReader(`{"id": "1", "name": "name"}`))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"id": "1", "name": "name"}`, string(body))

	list, err := store.List(BasicType)
	require.NoError(t, err)
	actual, err := json.Marshal(list)
	require.NoError(t, err)
	assert.JSONEq(t, `[{"id": "1", "name": "name"}]`, string(actual))
}

func TestServer_PostValidatesSchema(t *testing.T) {
	store, err := datastore.NewMemory()
	require.NoError(t, err)

	url, closeFn := startServer(store, BasicType)
	defer closeFn()

	body := strings.NewReader(`{"id": "1"}`)
	resp, err := http.Post(fmt.Sprintf("%s/%s", url, BasicType.Name), "application/json", body)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"validationErrors": ["(root): name is required"]}`, string(errBody))

	all2, err := store.List(BasicType)
	require.NoError(t, err)
	assert.Len(t, all2, 0)
}

func TestServer_Put(t *testing.T) {
	store, err := datastore.NewMemory()
	require.NoError(t, err)
	require.NoError(t, store.Create(BasicType, "1", server.Object{"id": 1, "value": "value1"}))

	url, closeFn := startServer(store, BasicType)
	defer closeFn()

	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s/%s", url, BasicType.Name, "1"), strings.NewReader(`{"id": "1", "name": "value2"}`))
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"id": "1", "name": "value2"}`, string(body))

	updated, err := store.Get(BasicType, "1")
	require.NoError(t, err)
	assert.Equal(t, "value2", updated["name"])
}

func TestServer_PutValidatesSchema(t *testing.T) {
	store, err := datastore.NewMemory()
	require.NoError(t, err)
	require.NoError(t, store.Create(BasicType, "1", server.Object{"id": 1, "name": "value1", "other": "first"}))

	url, closeFn := startServer(store, BasicType)
	defer closeFn()

	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s/%s", url, BasicType.Name, "1"), strings.NewReader(`{"id": "1", "other": "second"}`))
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"validationErrors": ["(root): name is required"]}`, string(errBody))

	updated, err := store.Get(BasicType, "1")
	require.NoError(t, err)
	assert.Equal(t, "first", updated["other"])
}

func TestServer_Delete(t *testing.T) {
	store, err := datastore.NewMemory()
	require.NoError(t, err)
	require.NoError(t, store.Create(BasicType, "1", server.Object{"id": 1, "value": "value1"}))

	url, closeFn := startServer(store, BasicType)
	defer closeFn()

	all, err := store.List(BasicType)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	request, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s/%s", url, BasicType.Name, "1"), nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	all2, err := store.List(BasicType)
	require.NoError(t, err)
	assert.Len(t, all2, 0)
}

func TestServer_DescribesContent(t *testing.T) {
	store, err := datastore.NewMemory()
	require.NoError(t, err)
	require.NoError(t, store.Create(BasicType, "1", server.Object{"id": 1, "value": "value1"}))

	url, closeFn := startServer(store, BasicType)
	defer closeFn()

	all, err := store.List(BasicType)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	resp, err := http.Get(fmt.Sprintf("%s/describe", url))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	errBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	r := describeResp{}
	err = json.Unmarshal(errBody, &r)
	require.NoError(t, err)

	assert.Len(t, r.Types, 1)
	assert.Equal(t, r.Types[0].Name, BasicType.Name)
	assert.Equal(t, r.Types[0].Id, BasicType.Id)
	assert.Equal(t, r.Types[0].Schema, BasicType.Schema)
}

type typeResp struct {
	Name   string `json:"name"`
	Id     string `json:"id"`
	Schema string `json:"schema"`
}
type describeResp struct {
	Types []typeResp `json:"types"`
}

func startServer(dataStore datastore.Memory, ty server.Type) (string, func()) {
	r := chi.NewRouter()
	server.Server{
		Config: server.Config{
			Types: []server.Type{ty},
		},
		DataStore: dataStore,
	}.Start(r)

	testServer := httptest.NewServer(r)
	return testServer.URL, testServer.Close
}
