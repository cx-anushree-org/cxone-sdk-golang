package aiscresults_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/checkmarx-open-labs/cxone-sdk-golang/internal/endpoints/aiscresults"
	"github.com/checkmarx-open-labs/cxone-sdk-golang/internal/retry"
	"github.com/checkmarx-open-labs/cxone-sdk-golang/internal/transport"
)

type stubAuth struct{}

func (stubAuth) Token(context.Context) (string, uint64, error)           { return "tok", 1, nil }
func (stubAuth) Refresh(context.Context, uint64) (string, uint64, error) { return "tok", 1, nil }

func newExec(t *testing.T, h http.HandlerFunc) (*transport.Executor, string) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	exec := transport.NewExecutor(transport.Config{
		HTTPClient:    srv.Client(),
		Authenticator: stubAuth{},
		UserAgent:     "test",
		CorrelationID: "corr",
		RetryPolicy:   retry.Policy{MaxAttempts: 1, MaxDelay: 0},
	})
	return exec, srv.URL + "/"
}

func TestList(t *testing.T) {
	var gotPath string
	exec, base := newExec(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = io.WriteString(w, `{
			"data": [
				{
					"id": "",
					"evidenceKey": "vulnerable_AI.py:11",
					"assetType": "AI SDK",
					"assetTypeId": "3b0bf814-e2ef-4603-b456-fa08c0b2e7eb",
					"assetId": "##anthropic##anthropic-python-sdk",
					"assetName": "Anthropic Python SDK",
					"provider": "Anthropic",
					"assetFirstDetectionDate": "2026-07-15T04:38:22Z",
					"path": "vulnerable_AI.py",
					"startLine": 11,
					"startColumn": 19,
					"endLine": 17,
					"endColumn": 46
				}
			],
			"total": 1,
			"currentPage": 1,
			"lastPage": 1
		}`)
	})

	out, err := aiscresults.List(context.Background(), exec, base, "scan-123", url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(gotPath, "/scans/scan-123/results") {
		t.Errorf("path = %q", gotPath)
	}
	if len(out.Data) != 1 || out.Data[0].AssetName != "Anthropic Python SDK" {
		t.Errorf("got %+v", out)
	}
	if out.Total != 1 || out.CurrentPage != 1 || out.LastPage != 1 {
		t.Errorf("pagination = %+v", out)
	}
}

func TestListValidation(t *testing.T) {
	exec, base := newExec(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach server")
	})
	_, err := aiscresults.List(context.Background(), exec, base, "", url.Values{})
	if err == nil {
		t.Error("expected error for empty scanID")
	}
}

func TestAggregate(t *testing.T) {
	var gotQuery string
	exec, base := newExec(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		if !strings.HasSuffix(r.URL.Path, "/scans/scan-123/results/aggregate") {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{
			"scanGroupsCounter": [
				{
					"assetTypeId": "2295665d-84ed-46d3-aecf-e57aa286bf13",
					"assetType": "AI Model",
					"provider": "OpenAI",
					"count": 3
				}
			]
		}`)
	})

	out, err := aiscresults.Aggregate(context.Background(), exec, base, "scan-123", "assetType,provider", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "groupBy=assetType%2Cprovider") {
		t.Errorf("query = %q", gotQuery)
	}
	if len(out.ScanGroupsCounter) != 1 || out.ScanGroupsCounter[0].Provider != "OpenAI" {
		t.Errorf("got %+v", out)
	}
	if out.ScanGroupsCounter[0].Count != 3 {
		t.Errorf("count = %d", out.ScanGroupsCounter[0].Count)
	}
}

func TestAggregateValidation(t *testing.T) {
	exec, base := newExec(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach server")
	})

	t.Run("empty scanID", func(t *testing.T) {
		_, err := aiscresults.Aggregate(context.Background(), exec, base, "", "assetType", nil)
		if err == nil {
			t.Error("expected error for empty scanID")
		}
	})

	t.Run("empty groupBy", func(t *testing.T) {
		_, err := aiscresults.Aggregate(context.Background(), exec, base, "scan-123", "", nil)
		if err == nil {
			t.Error("expected error for empty groupBy")
		}
	})
}