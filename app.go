package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var data map[string][]interface{}

var port = "3000"

type jsonServer struct {
	Status    string   `json:"status"`
	Endpoints []string `json:"endpoints"`
	Port      string   `json:"port"`
	Db        string   `json:db`
}

func main() {
	r := chi.NewRouter()
	endpoints := make([]string, 0)
	dbPath := "db.json"
	for _, v := range os.Args {
		if len(v) > 0 && strings.HasSuffix(v, "json") {
			dbPath = v
		}
	}

	byteValue, err := readJson(dbPath)
	if err != nil {
		panic(err)
		os.Exit(0)
	}

	jsonMarshalErr := json.Unmarshal([]byte(byteValue), &data)
	if jsonMarshalErr != nil {
		//TODO
	}
	for k, _ := range data {
		endpoints = append(endpoints, k)
		url := fmt.Sprintf("/%s", k)
		r.Mount(url, createEndpoints(k))
	}
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		serverInfo := &jsonServer{
			Endpoints: endpoints,
			Status:    "Running",
			Port:      port,
			Db:        dbPath}
		serverInfoMarshaled, err := json.Marshal(serverInfo)
		if err != nil {
			w.Write([]byte("Error"))
		}
		fmt.Fprint(w, string(serverInfoMarshaled))
	})

	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}

func createEndpoints(endpoint string) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			response, err := json.Marshal(data[endpoint])
			if err != nil {
				w.Write([]byte("Error"))
			}
			w.Write([]byte(response))
		})

		r.Post("/", createItem(endpoint))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", getItem)
			r.Delete("/", deleteItem(endpoint))
			r.Put("/", updateItem(endpoint))
		})
	})

	return r
}

func readJson(fileName string) ([]byte, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Successfully opened json %s", fileName)
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	return byteValue, nil
}

func getItem(endpoint string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		rawObj := data[endpoint]
		var result map[string]interface{}
		for _, v := range rawObj {
			item := v.(map[string]interface{})
			i, err := strconv.ParseFloat(id, 64)
			if err != nil {
				fmt.Println(err)
			}
			if item["id"] == i {
				result = item
				break
			}
		}
		res, _ := json.Marshal(result)
		w.Write(res)
	}
}

func deleteItem(endpoint string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		rawObj := data[endpoint]
		for k, v := range rawObj {
			item := v.(map[string]interface{})
			i, err := strconv.ParseFloat(id, 64)
			if err != nil {
				fmt.Println(err)
			}
			if item["id"] == i {
				arr := data[endpoint]
				arr = append(arr[:k], arr[k+1:]...)
				data[endpoint] = arr
				break
			}
		}
		res, _ := json.Marshal(data[endpoint])
		w.Write(res)
	}

}

func updateItem(endpoint string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		decoder := json.NewDecoder(r.Body)
		var body map[string]interface{}
		decoderErr := decoder.Decode(&body)
		if decoderErr != nil {
			w.Write([]byte("Error"))
		}
		rawObj := data[endpoint]
		for k, v := range rawObj {
			item := v.(map[string]interface{})
			i, err := strconv.ParseFloat(id, 64)
			if err != nil {
				fmt.Println(err)
			}
			if item["id"] == i {
				arr := data[endpoint]
				arr[k] = body
				data[endpoint] = arr
				break
			}
		}
		res, _ := json.Marshal(data[endpoint])
		w.Write(res)
	}
}

func createItem(endpoint string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var body map[string]interface{}
		decoderErr := decoder.Decode(&body)
		if decoderErr != nil {
			w.Write([]byte("Error"))
		}
		rawObj := data[endpoint]
		for _, v := range rawObj {
			item := v.(map[string]interface{})
			if item["id"] == body["id"] {
				w.Write([]byte("Error item with same id already exist! You can update instead of creating !"))
				return
			}
		}
		data[endpoint] = append(data[endpoint], body)
		res, _ := json.Marshal(data[endpoint])
		w.Write(res)
	}
}
