package app

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_normalizeEnvName(t *testing.T) {
	input := "Test-Camel-Dash-Name"
	expected := "TEST_CAMEL_DASH_NAME"

	output := normalizeEnvName(input)

	if output != expected {
		t.Errorf("Normalization failed. OUTPUT: %s, EXPECTED: %s", output, expected)
		return
	}
}

func Test_exportEnv_coockies(t *testing.T) {
	enableExportCookies := true
	enableExportHeaders := false
	envPrefix := "Web_Req_"
	req := httptest.NewRequest("GET", "localhost:20081", nil)

	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "session_id",
	})
	req.AddCookie(&http.Cookie{
		Name:  "x-token",
		Value: "token_value",
	})

	expected := []string{"WEB_PREFIX=WEB_REQ_", "WEB_REQ_COOKIE_SESSION=session_id", "WEB_REQ_COOKIE_X_TOKEN=token_value"}
	output := exportEnv(enableExportCookies, enableExportHeaders, envPrefix, req)

	if !reflect.DeepEqual(expected, output) {
		t.Errorf("Cookies export failed. OUTPUT: %v, EXPECTED: %v", output, expected)
		return
	}
}

func Test_exportEnv_headers(t *testing.T) {
	enableExportCookies := false
	enableExportHeaders := true
	envPrefix := "web_req_"
	req := httptest.NewRequest("GET", "localhost:20081", nil)

	req.Header.Add("Container-id", "ad313e67d06a")
	req.Header.Add("Role", "test-role")

	expected := []string{"WEB_PREFIX=WEB_REQ_", "WEB_REQ_HEADER_CONTAINER_ID=ad313e67d06a", "WEB_REQ_HEADER_ROLE=test-role"}
	output := exportEnv(enableExportCookies, enableExportHeaders, envPrefix, req)

	if len(expected) != len(output) {
		t.Errorf("Headers Len failed. OUTPUT: %v, EXPECTED: %v", output, expected)
		return
	}

	for _, e := range expected {
		if !stringInSlice(e, output) {
			t.Errorf("Headers export failed. v: %s, OUTPUT: %v, EXPECTED: %v", e, output, expected)
			return
		}
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
