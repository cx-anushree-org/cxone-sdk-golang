package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	cxone "github.com/checkmarx-open-labs/cxone-sdk-golang"
)

func main() {
	client, err := cxone.NewClient().
		Region(cxone.RegionIndia).
		Tenant(mustEnv("CX_TENANT")).
		AgentName("aisc-results-example").
		APIKey(mustEnv("CX_API_KEY")).
		Build()
	if err != nil {
		log.Fatalf("build client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	scanID := mustEnv("CX_SCAN_ID")

	// List
	q := url.Values{}
	q.Set("limit", "10")

	listOut, err := client.Advanced().AISCResults().List(ctx, scanID, q)
	if err != nil {
		log.Fatalf("list AISC results: %v", err)
	}
	fmt.Printf("Total findings: %d (page %d of %d)\n", listOut.Total, listOut.CurrentPage, listOut.LastPage)
	for _, r := range listOut.Data {
		fmt.Printf("  [%s] %s (%s) — %s:%d\n", r.AssetType, r.AssetName, r.Provider, r.Path, r.StartLine)
	}

	// Aggregate
	aggOut, err := client.Advanced().AISCResults().Aggregate(ctx, scanID, "assetType,provider", nil)
	if err != nil {
		log.Fatalf("aggregate AISC results: %v", err)
	}
	fmt.Println("\nAggregate by assetType,provider:")
	for _, g := range aggOut.ScanGroupsCounter {
		fmt.Printf("  %s / %s: %d\n", g.AssetType, g.Provider, g.Count)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}
