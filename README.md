
# CMS

Prototype of a lightweight CMS targeted towards prototyping and smaller projects.

Users define the types of entities they want to store and a CRUD REST API is automatically generated
for that schema, a React Admin UI is also created dynamically for those entities so users have a way to manage the data at runtime.

Data storage is customisable based on the data store chosen, currently in-memory and google clouds ECS object store are provided.

The aim is to have something that has no overhead to run, object storage is cheap any typically scales with the amount of data 
stored, the container starts quickly allowing scale to zero when not in use. Use this to build out as many different ideas
as you want and don't worry about database / hosting costs

## Roadmap

* Auth
* Caching
* Migration


## Getting started

The quickest way to get started is to run with the in-memory storage provider. Simply start the container 
mounting the config file from `examples/memory` like so.

```shell
docker run -p 8080:8080 \
  -v "$(pwd)"/examples/memory:/etc/cms \
  crswty/cms:latest
```

This will start the admin ui at `http://localhost:8080/admin` and the api at `http://localhost:8080/admin`.
You should be able to add/remove/edit records in the admin UI and see those changes in the api with:
```
curl http://localhost:8080/api/users
```

You can add/remove/customize the data types by configuring the `/examples/memory/config.yml` file you mounted.


## Data stores

### Google Cloud (ECS)

For an example of this see `config/gcloud`. Firstly configure the bucket name in `config.yml` and then
replace `credentials.json` with your credential file then run the following.

```shell
docker run -p 8080:8080 \
  -v "$(pwd)"/examples/gcloud:/etc/cms \
  crswty/cms:latest
```

## Development

Start the API
```shell
cd api
go run cmd/main.go
```

Start the UI
```shell
cd web
yarn start
```