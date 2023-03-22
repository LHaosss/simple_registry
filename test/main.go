package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.Encode("ssss")
	resp, err := http.Post("http://localhost:4000/update", "application/json", buf)
	fmt.Println(err)
	fmt.Println(resp)
}
