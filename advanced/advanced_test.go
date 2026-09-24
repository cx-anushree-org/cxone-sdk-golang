package advanced_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/checkmarx-open-labs/cxone-sdk-golang"
	"github.com/checkmarx-open-labs/cxone-sdk-golang/models"
)

// newClientAgainst spins up an httptest server that handles BOTH the IAM
// token endpoint AND the API endpoints, then constructs a real cxone.Client
// pointed at it via a custom region. This exercises the full builder →
// auth → executor → endpoint chain end-to-end.
func newClientAgainst(t *testing.T, api http.HandlerFunc) *cxone.Client {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/realms/acme/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"access_token":"tok-1","token_type":"Bearer","expires_in":300}`)
	})
	mux.HandleFunc("/api/", api)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	host := strings.TrimPrefix(srv.URL, "http://")
	region, err := cxone.NewCustomRegion(host, host)
	if err != nil {
		t.Fatal(err)
	}
	region, err = region.WithScheme("http")
	if err != nil {
		t.Fatal(err)
	}
	c, err := cxone.NewClient().
		Region(region).
		Tenant("acme").
		AgentName("test-agent").
		APIKey("dummy-refresh-token").
		HTTPClient(srv.Client()).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestAdvancedProjectsCreateEndToEnd(t *testing.T) {
	var (
		gotAuth   string
		gotAccept string
		gotCorrID string
		gotPath   string
	)
	c := newClientAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotCorrID = r.Header.Get("CorrelationId")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"p1","name":"acme"}`)
	})
	out, err := c.Advanced().Projects().Create(context.Background(), &models.Project{Name: "acme"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "p1" || out.Name != "acme" {
		t.Errorf("got %+v", out)
	}
	if gotAuth != "Bearer tok-1" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotAccept != "*/*; version=1.0" {
		t.Errorf("Accept = %q", gotAccept)
	}
	if gotCorrID != c.CorrelationID() {
		t.Errorf("CorrelationId mismatch: got %q want %q", gotCorrID, c.CorrelationID())
	}
	if gotPath != "/api/projects" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestAdvancedScansListEndToEnd(t *testing.T) {
	c := newClientAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/scans" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"totalCount":1,"filteredTotalCount":1,"scans":[{"id":"s1","status":"Completed"}]}`)
	})
	out, err := c.Advanced().Scans().List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Scans) != 1 || out.Scans[0].Status != models.ScanCompleted {
		t.Errorf("got %+v", out)
	}
}

func TestAdvancedUploadsEndToEnd(t *testing.T) {
	// We'll have the API server return a presigned URL pointing back to a
	// SECOND httptest server. The PUT to the second server must NOT carry
	// the bearer token.
	upSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("upload server method = %q", r.Method)
		}
		if r.Header.Get("Authorization") != "" {
			t.Errorf("upload server saw Authorization header: %q", r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "zipdata" {
			t.Errorf("upload body = %q", body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upSrv.Close()

	c := newClientAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/uploads" {
			t.Errorf("api path = %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"url":"`+upSrv.URL+`/u/abc"}`)
	})

	body := []byte("zipdata")
	url, err := c.Advanced().Uploads().UploadFromReader(context.Background(), bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, upSrv.URL) {
		t.Errorf("returned URL = %q", url)
	}
}

func TestAdvancedSCMDisconnectEndToEnd(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotAuth   string
	)
	c := newClientAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})
	err := c.Advanced().SCM().DisconnectProject(context.Background(), "proj-123")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/repos-manager/projects/proj-123/disconnect" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer tok-1" {
		t.Errorf("Authorization = %q", gotAuth)
	}
}

func TestAdvancedAISCResultsListEndToEnd(t *testing.T) {
	c := newClientAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ai-sc/reader/scans/scan-123/results" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
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
	out, err := c.Advanced().AISCResults().List(context.Background(), "scan-123", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Data) != 1 || out.Data[0].AssetName != "Anthropic Python SDK" {
		t.Errorf("got %+v", out)
	}
}

func TestAdvancedAISCResultsAggregateEndToEnd(t *testing.T) {
	c := newClientAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ai-sc/reader/scans/scan-123/results/aggregate" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("groupBy") != "assetType,provider" {
			t.Errorf("groupBy = %q", r.URL.Query().Get("groupBy"))
		}
		w.Header().Set("Content-Type", "application/json")
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
	out, err := c.Advanced().AISCResults().Aggregate(context.Background(), "scan-123", "assetType,provider", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.ScanGroupsCounter) != 1 || out.ScanGroupsCounter[0].Provider != "OpenAI" {
		t.Errorf("got %+v", out)
	}
}