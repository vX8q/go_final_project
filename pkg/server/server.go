package server

import (
	"fmt"
	"net/http"

	"go1f/pkg/api"
)

func Run(port string) error {

	api.Init()

	webDir := "./web"
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
