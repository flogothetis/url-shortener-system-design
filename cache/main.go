package main

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/bradfitz/gomemcache/memcache"
    "github.com/gorilla/mux"
)

var cache *memcache.Client

func init() {
    cache = memcache.New(ME_CONFIG_MEMCACHE_URL)
}

func getFromCacheHandler(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    key := params["key"]

    item, err := cache.Get(key)
    if err != nil {
        http.Error(w, "Key not found", http.StatusNotFound)
        return
    }

    response := map[string]string{key: string(item.Value)}
    json.NewEncoder(w).Encode(response)
}

func setToCacheHandler(w http.ResponseWriter, r *http.Request) {
    var data map[string]string
    err := json.NewDecoder(r.Body).Decode(&data)
    if err != nil {
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    key := data["key"]
    value := data["value"]

    err = cache.Set(&memcache.Item{Key: key, Value: []byte(value)})
    if err != nil {
        http.Error(w, "Failed to set value to cache", http.StatusInternalServerError)
        return
    }

    response := map[string]bool{"success": true}
    json.NewEncoder(w).Encode(response)
}

func main() {
    router := mux.NewRouter()

    router.HandleFunc("/get/{key}", getFromCacheHandler).Methods("GET")
    router.HandleFunc("/set", setToCacheHandler).Methods("POST")

    log.Fatal(http.ListenAndServe(":8080", router))
}
