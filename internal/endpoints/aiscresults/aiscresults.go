// Package aiscresults implements the Checkmarx One AI Supply Chain (AISC)
// results endpoints.

package aiscresults

import (
	"context"
	"net/http"
	"net/url"

	"github.com/checkmarx-open-labs/cxone-sdk-golang/cxerrors"
	"github.com/checkmarx-open-labs/cxone-sdk-golang/internal/transport"
	"github.com/checkmarx-open-labs/cxone-sdk-golang/models"
)

// ScansPath is the relative base path for AISC scan-scoped endpoints.
const ScansPath = "api/ai-sc/reader/scans"

// List returns a page of AI supply chain findings for the given scan.
// GET /api/ai-sc/reader/scans/{scanId}/results → 200.
func List(ctx context.Context, e *transport.Executor, baseURL, scanID string, query url.Values) (*models.AISCResultsCollection, error) {
	if scanID == "" {
		return nil, &cxerrors.ConfigurationError{Field: "scanID", Reason: "must not be empty"}
	}
	var out models.AISCResultsCollection
	if err := transport.DoJSON(ctx, e, http.MethodGet,
		transport.JoinURL(baseURL, ScansPath+"/"+scanID+"/results"),
		query, nil, []int{http.StatusOK}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Aggregate returns grouped counts of AI supply chain findings for the given scan.
// GET /api/ai-sc/reader/scans/{scanId}/results/aggregate → 200.
func Aggregate(ctx context.Context, e *transport.Executor, baseURL, scanID, groupBy string, query url.Values) (*models.AISCAggregateResponse, error) {
	if scanID == "" {
		return nil, &cxerrors.ConfigurationError{Field: "scanID", Reason: "must not be empty"}
	}
	if groupBy == "" {
		return nil, &cxerrors.ConfigurationError{Field: "groupBy", Reason: "must not be empty"}
	}
	q := query
	if q == nil {
		q = url.Values{}
	}
	q.Set("groupBy", groupBy)

	var out models.AISCAggregateResponse
	if err := transport.DoJSON(ctx, e, http.MethodGet,
		transport.JoinURL(baseURL, ScansPath+"/"+scanID+"/results/aggregate"),
		q, nil, []int{http.StatusOK}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}