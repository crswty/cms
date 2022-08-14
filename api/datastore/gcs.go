package datastore

import (
	"cloud.google.com/go/storage"
	"context"
	"crswty.com/cms/server"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"strings"
)

type Gcs struct {
	Client *storage.Client
	Bucket string
}

type GcsConfig struct {
	LocalTestUrl    *string
	Bucket          string
	CredentialsFile *string
}

func NewGcs(config GcsConfig) (Gcs, error) {
	opts := make([]option.ClientOption, 0)

	if config.LocalTestUrl != nil {
		opts = append(opts, option.WithEndpoint(*config.LocalTestUrl))
		opts = append(opts, option.WithoutAuthentication())
		opts = append(opts, option.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}}))
	}

	if config.CredentialsFile != nil {
		opts = append(opts, option.WithCredentialsFile(*config.CredentialsFile))
	}

	client, err := storage.NewClient(context.TODO(), opts...)
	if err != nil {
		return Gcs{}, fmt.Errorf("unable to create gcs client %w", err)
	}

	return Gcs{
		Client: client,
		Bucket: config.Bucket,
	}, nil
}

func (g Gcs) List(t server.Type) ([]server.Object, error) {
	objects := g.Client.Bucket(g.Bucket).Objects(context.TODO(), &storage.Query{
		Prefix: fmt.Sprintf("%s/", t.Name),
	})
	objs := make([]server.Object, 0)
	for {
		next, err := objects.Next()
		if err == iterator.Done {
			return objs, nil
		}
		if err != nil {
			return nil, fmt.Errorf("unable to list item: %w", err)
		}

		//TODO cache list responses based on ETag
		id := strings.TrimPrefix(next.Name, t.Name+"/") //TODO do this better
		get, err := g.Get(t, id)
		if err != nil {
			return nil, fmt.Errorf("unable to list item detail %s : %w", id, err)
		}
		objs = append(objs, get)
	}
}

func (g Gcs) Get(t server.Type, id string) (server.Object, error) {
	objectName := fmt.Sprintf("%s/%s", t.Name, id)
	object := g.Client.Bucket(g.Bucket).Object(objectName)
	reader, err := object.NewReader(context.TODO())
	if err != nil {
		return server.Object{}, fmt.Errorf("gcs provider failed to find  %s error: %w", objectName, err)
	}
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return server.Object{}, fmt.Errorf("gcs provider failed to read %s error: %w", objectName, err)
	}
	data := server.Object{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return server.Object{}, fmt.Errorf("gcs provider failed to unmarshal %s error: %w", objectName, err)
	}
	return data, nil
}

func (g Gcs) Create(t server.Type, id string, obj server.Object) error {
	object := g.Client.Bucket(g.Bucket).Object(fmt.Sprintf("%s/%s", t.Name, id))
	writer := object.NewWriter(context.TODO())
	marshal, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("gcs provider failed to create id %s error: %w", id, err)
	}
	_, err = writer.Write(marshal)
	if err != nil {
		return fmt.Errorf("gcs provider failed to create id %s error: %w", id, err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("gcs provider failed to create id %s error: %w", id, err)
	}

	return nil
}

func (g Gcs) Update(t server.Type, id string, obj server.Object) error {
	return g.Create(t, id, obj)
}

func (g Gcs) Delete(t server.Type, id string) error {
	err := g.Client.Bucket(g.Bucket).Object(fmt.Sprintf("%s/%s", t.Name, id)).Delete(context.TODO())
	if err != nil {
		return fmt.Errorf("gcs provider failed to delete id %s error: %w", id, err)
	}
	return nil
}
