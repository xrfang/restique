package main

import (
	"fmt"
	"net/http"
	"os"
)

func getpid(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, os.Getpid())
}
