package cachingClient

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"pokedex/internal/pokecache"
	"time"
)

var cache = pokecache.NewCache(20 * time.Second)

func Get(url string) ([]byte, error) {
	if res, err := http.Get(url); err == nil {
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if res.StatusCode > 299 {
			return nil, fmt.Errorf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal("Error reading body", err)
		}
		cache.Add(url, body)
		return body, nil
	} else {
		return []byte{}, err
	}
}

func CachedGet(url string) ([]byte, error) {
	if data, exists := cache.Get(url); exists {
		return data, nil
	}
	if res, err := http.Get(url); err == nil {
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal("Error reading body", err)
		}
		cache.Add(url, body)
		return body, nil
	} else {
		return []byte{}, err
	}
}
