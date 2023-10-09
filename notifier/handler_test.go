package notifier

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupSuite(tb testing.TB) func(t testing.TB) {
	cfg = config{
		Appdrift: phoneConfig{
			PhoneNumber: "+123",
			AccessToken: "apptoken",
		},
		InfrastrukturDrift: phoneConfig{
			PhoneNumber: "+345",
			AccessToken: "infratoken",
		},
	}

	// teardown
	return func(tb testing.TB) {
		cfg = config{}
	}
}
func TestNotify_should_send_sms_to_appdrift(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	message := SMSEagleMessage{
		Message:  "test",
		Receiver: Appdrift,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/http_api/send_sms" {
			t.Errorf("Expected to request '/http_api/send_sms'")
		}
		expectedQuery := fmt.Sprintf("access_token=apptoken&to=+123&message=test")
		if r.URL.RawQuery != expectedQuery {
			t.Errorf("Expected query %s, got %s", expectedQuery, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()
	cfg.Appdrift.Url = server.URL

	err := Notify(&message)

	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestNotify_should_send_sms_to_infradrift(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	message := SMSEagleMessage{
		Message:  "test",
		Receiver: Infrastruktur,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/http_api/send_sms" {
			t.Errorf("Expected to request '/http_api/send_sms'")
		}
		expectedQuery := fmt.Sprintf("access_token=infratoken&to=+345&message=test")
		if r.URL.RawQuery != expectedQuery {
			t.Errorf("Expected query %s, got %s", expectedQuery, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()
	cfg.InfrastrukturDrift.Url = server.URL

	err := Notify(&message)

	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestNotify_should_send_sms_and_call_infradrift(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	message := SMSEagleMessage{
		Message:  "test",
		Receiver: Infrastruktur,
		Call:     true,
	}

	mux := http.NewServeMux()

	//check sms
	mux.HandleFunc("/http_api/send_sms", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/http_api/send_sms" {
			t.Errorf("Expected to request '/http_api/send_sms'")
		}
		expectedQuery := fmt.Sprintf("access_token=infratoken&to=+345&message=test")
		if r.URL.RawQuery != expectedQuery {
			t.Errorf("Expected query %s, got %s", expectedQuery, r.URL.RawQuery)
		}
	})

	//check call

	mux.HandleFunc("/http_api/call_with_termination", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/http_api/call_with_termination" {
			t.Errorf("Expected to request '/http_api/call_with_termination'")
		}
		expectedQuery := fmt.Sprintf("access_token=infratoken&to=+345")
		if r.URL.RawQuery != expectedQuery {
			t.Errorf("Expected query %s, got %s", expectedQuery, r.URL.RawQuery)
		}
	})
	server := httptest.NewServer(mux)

	defer server.Close()
	cfg.InfrastrukturDrift.Url = server.URL

	err := Notify(&message)

	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}
