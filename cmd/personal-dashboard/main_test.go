package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(body string) *http.Response {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}
}
func TestDeepseekMultipleCurrenciesAndZero(t *testing.T) {
	a := app{client: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		return response(`{"is_available":true,"balance_infos":[{"currency":"CNY","total_balance":"0","granted_balance":"0","topped_up_balance":"0"},{"currency":"USD","total_balance":"5","granted_balance":"2","topped_up_balance":"3"}]}`), nil
	})}}
	m := a.deepseek(context.Background(), "fixture")
	if m.State != "ok" || len(m.Lines) != 4 {
		t.Fatalf("%+v", m)
	}
}
func TestOpenAIPagination(t *testing.T) {
	n := 0
	a := app{client: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		n++
		if n == 1 {
			return response(`{"data":[{"results":[{"amount":{"value":2,"currency":"usd"}}]}],"has_more":true,"next_page":"next"}`), nil
		}
		if r.URL.Query().Get("page") != "next" {
			t.Error("cursor missing")
		}
		return response(`{"data":[{"results":[{"amount":{"value":3,"currency":"usd"}}]}],"has_more":false}`), nil
	})}}
	m := a.openai(context.Background(), "fixture")
	if n != 2 || !strings.Contains(m.Lines[0], "5.00") {
		t.Fatalf("%+v", m)
	}
}
func TestSubscriptionsDoNotInventZero(t *testing.T) {
	m := parseSubscriptions([]byte(fmt.Sprintf(`{"providers":[{"id":"codex","updatedAt":%q,"windows":[{"kind":"session"}]}]}`, time.Now().Format(time.RFC3339))))
	if !strings.Contains(m.Lines[0], "未知") {
		t.Fatal(m)
	}
}
func TestSubscriptionsRejectStale(t *testing.T) {
	m := parseSubscriptions([]byte(`{"providers":[{"id":"claude","updatedAt":"2020-01-01T00:00:00Z","windows":[{"remainingPercent":90}]}]}`))
	if !strings.Contains(m.Lines[0], "过期") {
		t.Fatal(m)
	}
}
func TestSettingsDoNotExposeSecretsAndRejectCrossSite(t *testing.T) {
	a := app{config: settings{DeepSeek: "secret"}, token: "token"}
	w := httptest.NewRecorder()
	a.settings(w, httptest.NewRequest("GET", "/api/settings", nil))
	if strings.Contains(w.Body.String(), "secret") {
		t.Fatal("secret exposed")
	}
	w = httptest.NewRecorder()
	a.settings(w, httptest.NewRequest("POST", "/api/settings", strings.NewReader(`{}`)))
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
}

func TestMissingCostsNeverBecomeZero(t *testing.T) {
	a := app{client: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { return response(`{}`), nil })}}
	m := a.openai(context.Background(), "fixture")
	if m.State != "error" {
		t.Fatal(m)
	}
}
func TestAuthorizationErrorIsNotBalance(t *testing.T) {
	a := app{client: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		v := response(`{"error":"private detail"}`)
		v.StatusCode = 401
		return v, nil
	})}}
	m := a.deepseek(context.Background(), "fixture")
	if m.State != "error" || strings.Contains(strings.Join(m.Lines, ""), "private") {
		t.Fatal(m)
	}
}
