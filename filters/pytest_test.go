package filters

import (
	"strings"
	"testing"
)

func TestFilterPytest_WithFailures(t *testing.T) {
	raw := `============================= test session starts ==============================
platform linux -- Python 3.11.0, pytest-7.4.0, pluggy-1.3.0
rootdir: /home/user/project
collected 45 items

tests/test_auth.py ....F..                                              [ 15%]
tests/test_api.py ...........                                            [ 40%]
tests/test_utils.py ................                                     [100%]

=================================== FAILURES ===================================
_________________________________ test_login __________________________________

    def test_login():
>       assert response.status_code == 200
E       AssertionError: assert 401 == 200

tests/test_auth.py:25: AssertionError
=========================== short test summary info ============================
FAILED tests/test_auth.py::test_login - AssertionError: assert 401 == 200
========================= 1 failed, 44 passed in 3.2s =========================`

	got, err := filterPytest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) >= len(raw) {
		t.Errorf("expected compression, got %d >= %d", len(got), len(raw))
	}
	if !strings.Contains(got, "FAILED tests/test_auth.py") {
		t.Error("expected failed test name")
	}
	if !strings.Contains(got, "1 failed, 44 passed") {
		t.Error("expected summary")
	}
}

func TestFilterPytest_AllPass(t *testing.T) {
	raw := `============================= test session starts ==============================
platform linux -- Python 3.11.0, pytest-7.4.0
collected 20 items

tests/test_all.py ....................                                    [100%]

============================== 20 passed in 2.1s ===============================`

	got, err := filterPytest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "20 passed in 2.1s") {
		t.Errorf("expected pass summary, got: %s", got)
	}
}

func TestFilterPytest_Empty(t *testing.T) {
	got, err := filterPytest("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
