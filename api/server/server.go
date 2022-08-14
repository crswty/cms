package server

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/xeipuuv/gojsonschema"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	Config    Config
	DataStore DataProvider
}
type Type struct {
	Name   string
	Id     string
	Schema string
}
type Config struct {
	Types       []Type
	Schema      string
	AdminAssets string
}

type Object map[string]interface{}

type DataProvider interface {
	List(t Type) ([]Object, error)
	Get(t Type, id string) (Object, error)
	Create(t Type, id string, obj Object) error
	Update(t Type, id string, obj Object) error
	Delete(t Type, id string) error
}

func (s Server) Start(r chi.Router) {
	config := s.Config

	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: log.Default()}))

	r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
	})

	r.Route("/api", func(r chi.Router) {
		for _, t := range config.Types {
			s.addEndpoints(r, t)
		}

		r.Get("/describe", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")

			resp := make([]typeResp, 0)
			for _, t := range config.Types {
				resp = append(resp, typeResp{
					Name:   t.Name,
					Id:     t.Id,
					Schema: t.Schema,
				})
			}

			data, err := json.Marshal(describeResp{Types: resp})

			if err != nil {
				handleError(writer, err)
			}

			_, err = writer.Write(data)
			if err != nil {
				handleError(writer, err)
				return
			}
		})
	})

	contentDir := config.AdminAssets
	fs := http.FileServer(http.Dir(contentDir))
	r.Get("/admin", func(writer http.ResponseWriter, request *http.Request) {
		http.StripPrefix("/admin", fs).ServeHTTP(writer, request)
	})
	r.Get("/admin/*", func(writer http.ResponseWriter, request *http.Request) {
		if _, err := os.Stat(contentDir + strings.TrimPrefix(request.RequestURI, "/admin")); os.IsNotExist(err) {
			http.StripPrefix(request.RequestURI, fs).ServeHTTP(writer, request)
		} else {
			http.StripPrefix("/admin", fs).ServeHTTP(writer, request)
		}
	})

}

func (s Server) addEndpoints(r chi.Router, t Type) {
	r.Get(fmt.Sprintf("/%s", t.Name), func(writer http.ResponseWriter, request *http.Request) {
		data, err := s.DataStore.List(t)
		if err != nil {
			handleError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-Total-Count", fmt.Sprintf("%d", len(data)))

		b, err := json.Marshal(data)
		if err != nil {
			handleError(writer, err)
			return
		}
		_, err = writer.Write(b)
		if err != nil {
			handleError(writer, err)
			return
		}
	})

	r.Get(fmt.Sprintf("/%s/{id}", t.Name), func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		id := chi.URLParam(request, "id")
		data, err := s.DataStore.Get(t, id)
		if err != nil {
			handleError(writer, err)
			return
		}

		b, err := json.Marshal(data)
		if err != nil {
			handleError(writer, err)
			return
		}
		_, err = writer.Write(b)
		if err != nil {
			handleError(writer, err)
			return
		}
	})

	r.Post(fmt.Sprintf("/%s", t.Name), func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		reqBytes, err := io.ReadAll(request.Body)
		if err != nil {
			handleError(writer, err)
			return
		}

		valid, validationErrors, err := validate(t.Schema, string(reqBytes))
		if err != nil {
			handleError(writer, err)
			return
		}
		if !valid {
			handleValidationError(writer, validationErrors)
			return
		}

		var obj Object
		err = json.Unmarshal(reqBytes, &obj)
		if err != nil {
			handleError(writer, err)
			return
		}

		idStr, err := idToString(obj[t.Id])
		if err != nil {
			handleError(writer, err)
			return
		}
		err = s.DataStore.Create(t, idStr, obj)
		if err != nil {
			handleError(writer, err)
			return
		}

		writer.WriteHeader(http.StatusCreated)
		_, err = writer.Write(reqBytes)
		if err != nil {
			handleError(writer, err)
			return
		}
	})

	r.Put(fmt.Sprintf("/%s/{id}", t.Name), func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		id := chi.URLParam(request, "id")
		reqBytes, err := io.ReadAll(request.Body)
		if err != nil {
			handleError(writer, err)
			return
		}

		var obj Object
		err = json.Unmarshal(reqBytes, &obj)
		if err != nil {
			handleError(writer, err)
			return
		}

		valid, validationErrors, err := validate(t.Schema, string(reqBytes))
		if err != nil {
			handleError(writer, err)
			return
		}
		if !valid {
			handleValidationError(writer, validationErrors)
			return
		}

		err = s.DataStore.Update(t, id, obj)
		if err != nil {
			handleError(writer, err)
			return
		}

		writer.WriteHeader(http.StatusOK)
		_, err = writer.Write(reqBytes)
		if err != nil {
			handleError(writer, err)
			return
		}
	})

	r.Delete(fmt.Sprintf("/%s/{id}", t.Name), func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		id := chi.URLParam(request, "id")

		err := s.DataStore.Delete(t, id)
		if err != nil {
			handleError(writer, err)
			return
		}

		writer.WriteHeader(http.StatusNoContent)
	})
}

func validate(schema, content string) (bool, []string, error) {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	contentLoader := gojsonschema.NewStringLoader(content)
	result, err := gojsonschema.Validate(schemaLoader, contentLoader)
	if err != nil {
		return false, nil, err
	}

	errs := make([]string, 0)
	for _, resultError := range result.Errors() {
		errs = append(errs, resultError.String())
	}

	return result.Valid(), errs, nil
}

type errorResp struct {
	Message string `json:"message"`
}
type validationErrorResp struct {
	ValidationErrors []string `json:"validationErrors"`
}

func handleValidationError(writer http.ResponseWriter, messages []string) {
	log.Printf("Validation error: %s \n", messages)
	body, err := json.Marshal(validationErrorResp{ValidationErrors: messages})
	if err != nil {
		log.Printf("error handling valiation error: %s \n", err)
		return
	}
	writer.WriteHeader(http.StatusBadRequest)
	_, _ = writer.Write(body)
}

func handleError(writer http.ResponseWriter, e error) {
	log.Printf("Error: %s \n", e)
	body, err := json.Marshal(errorResp{Message: e.Error()})
	if err != nil {
		log.Printf("error handling error: %s \n", err)
		return
	}
	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = writer.Write(body)
}

type describeResp struct {
	Types []typeResp `json:"types"`
}

type typeResp struct {
	Name   string `json:"name"`
	Id     string `json:"id"`
	Schema string `json:"schema"`
}

func idToString(id any) (string, error) {
	switch t := id.(type) {
	case string:
		return t, nil
	case int:
		return strconv.Itoa(t), nil
	default:
		return "", fmt.Errorf("invalid id type: %v", id)
	}
}
