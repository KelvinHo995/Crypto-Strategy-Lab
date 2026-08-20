package httpx

import "net/http"

// TODO: wire to strategy.Registry once it has a List method.

func listStrategies(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
