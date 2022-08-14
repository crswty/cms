package main

import (
	"crswty.com/cms/datastore"
	"crswty.com/cms/server"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
	"net/http"
)

const appName = "cms"

func main() {
	//TODO work out how to handle IDs
	//TODO validate schema on start
	v := viper.New()

	v.AddConfigPath(fmt.Sprintf("/etc/%s", appName))
	v.AddConfigPath(".")
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("unable to read config: %w", err))
	}

	typesFromConfig, err := getTypesFromConfig(v)
	if err != nil {
		panic(fmt.Errorf("unable to parse type config: %w", err))
	}

	store, err := getProviderFromConfig(v, typesFromConfig)
	if err != nil {
		panic(fmt.Errorf("unable to create provider: %w", err))
	}

	v.SetDefault("adminAssets", "./web")

	r := chi.NewRouter()
	server.Server{
		Config: server.Config{
			Types:       typesFromConfig,
			AdminAssets: v.GetString("adminAssets"),
		},
		DataStore: store,
	}.Start(r)

	//TODO PORT var
	port := "8080"
	fmt.Printf("server starting at localhost:%s\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", "8080"), r)
	if err != nil {
		panic(err)
	}
}

type typeConfig []struct {
	Name   string `json:"name"`
	Id     string `json:"id"`
	Schema string `json:"schema"`
}

func getTypesFromConfig(v *viper.Viper) ([]server.Type, error) {
	types := typeConfig{}
	err := v.UnmarshalKey("types", &types)
	if err != nil {
		return nil, err
	}

	var ts = make([]server.Type, 0)
	for _, t := range types {
		ts = append(ts, server.Type{Name: t.Name, Id: t.Id, Schema: t.Schema})
	}
	return ts, nil
}

func getProviderFromConfig(v *viper.Viper, typesFromConfig []server.Type) (server.DataProvider, error) {
	var (
		store server.DataProvider
		err   error
	)

	providerName := v.GetString("provider.name")
	switch providerName {
	case "memory":
		store, err = getMemoryProvider(v, typesFromConfig)
	case "gcs":
		store, err = getGcsProvider(v)
	default:
		err = fmt.Errorf("no provider found with name: %s", providerName)
	}

	return store, err
}

func getGcsProvider(v *viper.Viper) (server.DataProvider, error) {
	credentialsFile := v.GetString("provider.credentialsFile")
	bucket := v.GetString("provider.bucket")
	return datastore.NewGcs(datastore.GcsConfig{
		Bucket:          bucket,
		CredentialsFile: &credentialsFile,
	})
}

type memoryProviderOptions struct {
	Name        string `json:"name"`
	InitialData []struct {
		Type string                 `json:"type"`
		Id   string                 `json:"id"`
		Data map[string]interface{} `json:"data"`
	} `json:"initialData"`
}

func getMemoryProvider(v *viper.Viper, typesFromConfig []server.Type) (server.DataProvider, error) {
	providerOptions := memoryProviderOptions{}
	err := v.UnmarshalKey("provider", &providerOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to parse memory provider options: %w", err)
	}

	records := make([]datastore.Record, 0)
	for _, datum := range providerOptions.InitialData {
		typeToAdd := typeForName(typesFromConfig, datum.Type)
		if typeToAdd == nil {
			return nil, fmt.Errorf("could not find registerd type for insert data %+v", datum)
		}

		records = append(records, datastore.Record{Id: datum.Id, Type: *typeToAdd, Data: datum.Data})
	}

	return datastore.NewMemory(records...)
}

func typeForName(types []server.Type, name string) *server.Type {
	for _, t := range types {
		if t.Name == name {
			return &t
		}
	}
	return nil
}
