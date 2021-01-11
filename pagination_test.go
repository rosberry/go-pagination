package pagination

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMainFlow(t *testing.T) {
	//http request-response
	ts := httptest.NewServer(setupServer())
	defer ts.Close()

	// Make a request to our server with the {base url}/ping
	resp, err := http.Get(fmt.Sprintf("%s/list", ts.URL))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

}

//controller

//model
