package service

import (
	"net/http"
	"testing"
)

func TestIsUpstreamBillingError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   []byte
		want   bool
	}{
		{name: "payment required status", status: http.StatusPaymentRequired, want: true},
		{name: "credit balance message", status: http.StatusBadRequest, body: []byte(`{"error":{"message":"credit balance is too low"}}`), want: true},
		{name: "insufficient balance message", status: http.StatusBadGateway, body: []byte(`{"message":"insufficient balance"}`), want: true},
		{name: "ordinary bad request", status: http.StatusBadRequest, body: []byte(`{"error":{"message":"invalid request"}}`), want: false},
		{name: "ordinary server error", status: http.StatusBadGateway, body: []byte(`{"message":"temporary upstream failure"}`), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUpstreamBillingError(tt.status, tt.body); got != tt.want {
				t.Fatalf("IsUpstreamBillingError() = %v, want %v", got, tt.want)
			}
		})
	}
}
