package miidd

import (
	"context"
	"net/http"

	"github.com/zacksfF/FullStack-Blockchain/web2"
)

// Cors sets the response headers needed for Cross-Origin Resource Sharing
func Cors(origin string) web2.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web2.Handler) web2.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// Set the CORS headers to the response.
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			// Call the next handler.
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
