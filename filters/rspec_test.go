package filters

import (
	"strings"
	"testing"
)

func TestFilterRspec(t *testing.T) {
	raw := `Randomized with seed 12345

AppController
  GET /index
    returns a success response
  POST /create
    creates a new item (FAILED - 1)

Failures:

  1) AppController POST /create creates a new item
     Failure/Error: expect(response).to have_http_status(:created)
     # ./spec/controllers/app_controller_spec.rb:25

Finished in 3.5 seconds (files took 2.1 seconds to load)
15 examples, 1 failure

Failed examples:

rspec ./spec/controllers/app_controller_spec.rb:20 # AppController POST /create creates a new item`

	got, err := filterRspec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression")
	}
	if !strings.Contains(got, "15 examples, 1 failure") {
		t.Error("expected summary")
	}
	if !strings.Contains(got, "rspec ./spec") {
		t.Error("expected failed example")
	}
}

func TestFilterRspec_Empty(t *testing.T) {
	got, err := filterRspec("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
