package datastore_test

import (
	"crswty.com/cms/datastore"
	"crswty.com/cms/server"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_MemoryStoreFulfilsContract(t *testing.T) {
	memory, err := datastore.NewMemory()
	require.NoError(t, err)
	Contract(t, memory)
}

func Test_GcsStoreFulfilsContract(t *testing.T) {
	s, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		Port:       8081,
		PublicHost: "localhost:8081", // see: https://github.com/fsouza/fake-gcs-server/issues/201
	})
	require.NoError(t, err)

	s.CreateBucketWithOpts(fakestorage.CreateBucketOpts{
		Name:              "cms-test-bucket",
		VersioningEnabled: true,
	})

	url := "localhost:8081"
	store, err := datastore.NewGcs(datastore.GcsConfig{
		LocalTestUrl: &url,
		Bucket:       "cms-test-bucket",
	})
	require.NoError(t, err)

	Contract(t, store)
}

func Contract(t *testing.T, provider server.DataProvider) {

	usersType := server.Type{Name: "users", Id: "id", Schema:
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

	petsType := server.Type{Name: "pets", Id: "id", Schema:
	// language=json
	`{
	"type": "object",
	"required": ["id"],
	"properties": { "id": { "type": "string" } }
}`}

	user1 := server.Object{"id": "1", "name": "value1"}
	user2 := server.Object{"id": "2", "name": "value2"}
	user3 := server.Object{"id": "3", "name": "value3"}

	pet1 := server.Object{"id": "1"}

	t.Run("create", func(t *testing.T) {
		assert.NoError(t, provider.Create(usersType, "1", user1))
		assert.NoError(t, provider.Create(usersType, "2", user2))
		assert.NoError(t, provider.Create(usersType, "3", user3))

		assert.NoError(t, provider.Create(petsType, "1", pet1))
	})

	t.Run("list", func(t *testing.T) {
		list, err := provider.List(usersType)
		require.NoError(t, err)
		assert.Len(t, list, 3)
		assert.Contains(t, list, user1)
		assert.Contains(t, list, user2)
		assert.Contains(t, list, user3)

	})

	t.Run("get", func(t *testing.T) {
		obj, err := provider.Get(usersType, "2")
		require.NoError(t, err)
		assert.Equal(t, user2, obj)
	})

	t.Run("update", func(t *testing.T) {
		err := provider.Update(usersType, "2", server.Object{"id": 2, "name": "updatedValue2"})
		require.NoError(t, err)

		obj, err := provider.Get(usersType, "2")
		require.NoError(t, err)
		assert.Equal(t, "updatedValue2", obj["name"])
	})

	t.Run("delete", func(t *testing.T) {
		err := provider.Delete(usersType, "2")
		require.NoError(t, err)

		list, err := provider.List(usersType)
		require.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Contains(t, list, user1)
		assert.Contains(t, list, user3)
	})
}
