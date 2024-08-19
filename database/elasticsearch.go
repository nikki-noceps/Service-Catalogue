package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"nikki-noceps/serviceCatalogue/config"
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/logger/tag"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type ESClient struct {
	client *es.Client
}

// Creates a new elasticsearch client. Tests the connection and returns it.
// Returns a wrapped error in case of any issues
func InitESClient(cfg config.ElasticSearch) (*ESClient, error) {
	esConfig := es.Config{
		Addresses: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	es, err := es.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create new elasticsearch client: %w", err)
	}

	info, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to ping elasticsearch: %w", err)
	}
	defer info.Body.Close()

	return &ESClient{
		client: es,
	}, nil
}

// searchRequest is a driver function which returns a raw response from elasticsearch.
// It searches using query provided. Lookup Query struct for supported operations.
func (es *ESClient) searchRequest(cctx context.CustomContext, query *Body, index string) (*esapi.Response, error) {
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(queryBytes),
	}
	res, err := req.Do(cctx, es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute es request: %w", err)
	}
	return res, nil
}

func (es *ESClient) handleSearchResponse(cctx context.CustomContext, res *esapi.Response) ([]any, error) {
	defer res.Body.Close()

	if res.IsError() {
		cctx.Logger().ERROR("search failed", tag.NewAnyTag("res", res))
		var e map[string]any
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("error parsing the response body: %w", err)
		}
		return nil, fmt.Errorf("search failed, got [%s] with error: %v", res.Status(), e)
	}

	var r map[string]any
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %w", err)
	}

	hits, ok := r["hits"].(map[string]interface{})["hits"].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to parse hits in es response")
	}

	return hits, nil
}

func (es *ESClient) SearchAndGetHits(cctx context.CustomContext, query *Body, index string) ([]any, error) {
	res, err := es.searchRequest(cctx, query, index)
	if err != nil {
		return nil, err
	}

	hits, err := es.handleSearchResponse(cctx, res)
	if err != nil {
		return nil, err
	}
	return hits, nil
}

// CreateDocument takes in bytes of body to create document in index provided
// Returns create respones and error if any
func (es *ESClient) CreateDocument(cctx context.CustomContext, docBytes []byte, index string) (*esapi.Response, error) {
	req := esapi.IndexRequest{
		Index: index,
		Body:  bytes.NewReader(docBytes),
	}

	// Execute the request
	res, err := req.Do(cctx, es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute es request: %w", err)
	}

	// Handle the response
	if res.IsError() {
		var e map[string]any
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("error parsing the response body: %w", err)
		}
		return nil, fmt.Errorf("create failed, got [%s] status code %v", res.Status(), e)
	}

	defer res.Body.Close()

	return res, nil
}

// UpdateDocument takes in bytes to be replaced for the documentId provided. It only updates parts of the document given in inputs
// Does not update rest of the fields which are not provided.
func (es *ESClient) UpdateDocument(cctx context.CustomContext, docBytes []byte, index string, docId string) (*esapi.Response, error) {
	res, err := es.client.Update(index, docId, bytes.NewReader(docBytes), es.client.Update.WithContext(cctx))
	if err != nil {
		cctx.Logger().DEBUG("failed to update document", tag.NewErrorTag(err))
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		cctx.Logger().DEBUG("updated failed", tag.NewAnyTag("res", res.String()))
		var e map[string]any
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("error parsing the response body: %w", err)
		}
		return nil, fmt.Errorf("update failed, got [%s] status code %v", res.Status(), e)
	}

	return res, nil
}
