# jwks-federation-server

A simple HTTP service which serves a JWKS endpoint returning public keys which are fetched from multiple upstream sources.

## Usage

Create a config file named `config.yaml` based on the sample `config.sample.yaml` file. Next, update `jwks_urls` and optionally limit the keys which gets import from upstream JWKS URLs by defining `allowed_kids`. 

```sh
# start container
podman run -d -p 8080:8080 -v $PWD/config.yaml:/config.yaml:ro ghcr.io/nimbolus/jwks-federation-server
# sample request
curl localhost:8080/.well-known/jwks.json
```
