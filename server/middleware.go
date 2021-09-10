package server

import (
	"context"
	"errors"
	"net/http"
)

type ContextKey string

const ContextPublicKey ContextKey = "publicKey"

// Order of the Middleware
// 1. set headers
// 2. check Cors
// 3. check Limits Rates
// 4. check Allowed Ip Ranges
// 5. preflight
// 6. Check Permissions

// SetHeaders // prepare header response
func SetHeaders() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			(w).Header().Set("Content-Type", "application/json")
			(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-JWT, Authorization")

			// continue
			h.ServeHTTP(w, r)
		})
	}
}

// enable cors within the http handler
func CheckCors(meta *Meta) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			allowed := false
			origin := r.Header.Get("Origin")

			// check if cors is available in list
			for _, v := range meta.Config.Server.AllowedOrigins {
				if v == origin {
					allowed = true
				}
			}

			// allow cors if found
			if !allowed {
				(w).Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				(w).Header().Set("Access-Control-Allow-Origin", origin)
			}

			// continue
			h.ServeHTTP(w, r)
		})
	}
}

// CheckLimitsRates handle limits and rates
func CheckLimitsRates() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// limit us requests per second
			limiter := limiter.GetLimiter(limiter.getIP(r))
			if !limiter.Allow() {
				(w).WriteHeader(http.StatusTooManyRequests)
				return
			}

			// continue
			h.ServeHTTP(w, r)
		})
	}
}

// CheckAllowedIPRange allow specific IP address
func CheckAllowedIPRange(meta *Meta) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			allowed := false
			ip := limiter.getIP(r)
			// check if ip is available in list
			for _, v := range meta.Config.Server.AllowedOrigins {
				if v == ip {
					allowed = true
				}
			}
			// allow ip if found
			if !allowed {
				(w).WriteHeader(http.StatusForbidden)
				return
			}

			// continue
			h.ServeHTTP(w, r)
		})
	}
}

// CheckAuth validates request for jwt header
func Preflight() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if (*r).Method == "OPTIONS" {
				(w).WriteHeader(http.StatusOK)
				return
			}

			// continue
			h.ServeHTTP(w, r)
		})
	}
}

// CheckAuth validates request for jwt header
func CheckAuth(meta *Meta) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// validate jwt
			err := meta.JWT.TokenValid(r)
			if err != nil {
				Error(w, http.StatusUnauthorized, errors.New("Unauthorized"), meta.Config.Server.Encryption.PublicKey)
				return
			}

			// continue
			ctx := context.WithValue(r.Context(), ContextPublicKey, meta.Config.Server.Encryption.PublicKey)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithoutAuth access without authentications
func WithoutAuth() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// continue
			h.ServeHTTP(w, r)
		})
	}
}

// CheckPermission validate if users is allowed to access route
func CheckPermission(meta *Meta) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var allowed bool
			var pattern = r.URL.Path
			token, err := meta.JWT.GetToken(r)
			if err != nil {
				Error(w, http.StatusUnauthorized, errors.New("Unauthorized"), meta.Config.Server.Encryption.PublicKey)
				return
			}

			for _, value := range token.Permissions {
				if value == pattern {
					allowed = true
				}
			}

			if allowed {
				Error(w, http.StatusUnauthorized, errors.New("Unauthorized"), meta.Config.Server.Encryption.PublicKey)
				return
			}

			// continue
			h.ServeHTTP(w, r)
		})
	}
}

// Middleware strct
type Middleware func(http.Handler) http.Handler

// Use middleware
func Use(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
