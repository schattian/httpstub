package httpstub

import "net/http"

// Client is the minimum interface needed to intercept requests for a http.Client obj
// Specially good fit with structs wrapping http.Client
type Client interface {
	SetClient(http.Client)
}
