package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	// for RSA keys
	N string `json:"n,omitempty"`
	E string `json:"e,omitempty"`
	// for EC keys
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

var (
	allowedKids map[string]bool
	cachedJWKS  JWKS
	mutex       sync.RWMutex
)

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("jwks_federation")
	viper.AutomaticEnv()

	viper.SetDefault("upstream_jwks_urls", []string{})
	viper.SetDefault("allowed_kids", []string{})
	viper.SetDefault("update_interval", "1h")
	viper.SetDefault("listen_addr", ":8080")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	// Initialize allowedKids map
	allowedKids = make(map[string]bool)
	for _, kid := range viper.GetStringSlice("allowed_kids") {
		allowedKids[kid] = true
	}
}

func fetchJWKS(url string) (JWKS, error) {
	resp, err := http.Get(url)
	if err != nil {
		return JWKS{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return JWKS{}, err
	}

	var jwks JWKS
	err = json.Unmarshal(body, &jwks)
	if err != nil {
		return JWKS{}, err
	}

	return jwks, nil
}

func updateJWKS() {
	var allKeys []JWK
	for _, url := range viper.GetStringSlice("upstream_jwks_urls") {
		jwks, err := fetchJWKS(url)
		if err != nil {
			fmt.Printf("Error fetching JWKS from %s: %v\n", url, err)
			continue
		}
		for _, key := range jwks.Keys {
			if len(allowedKids) == 0 || allowedKids[key.Kid] {
				allKeys = append(allKeys, key)
			}
		}
	}

	mutex.Lock()
	cachedJWKS.Keys = allKeys
	mutex.Unlock()
}

func jwksHandler(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cachedJWKS)
}

func main() {
	initConfig()

	updateJWKS()
	go func() {
		updateInterval, err := time.ParseDuration(viper.GetString("update_interval"))
		if err != nil {
			panic(fmt.Sprintf("Invalid update_interval: %v", err))
		}
		for {
			time.Sleep(updateInterval)
			updateJWKS()
		}
	}()

	http.HandleFunc("/.well-known/jwks.json", jwksHandler)
	addr := viper.GetString("listen_addr")
	fmt.Printf("Server starting on %s\n", addr)
	http.ListenAndServe(addr, nil)
}
