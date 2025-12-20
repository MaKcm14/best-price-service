package chttp

import (
	"strings"

	"github.com/MaKcm14/price-service/pkg/entities"
)

// ControllerConfig defines the controller's configuration.
type ControllerConfig struct {
	AsyncMode bool
}

// ProductResponse defines the response data.
type ProductResponse struct {
	Samples map[string]entities.ProductSample `json:"samples"`
}

func NewProductResponse(samples []entities.ProductSample) ProductResponse {
	marketSamples := make(map[string]entities.ProductSample)

	for _, sample := range samples {
		marketSamples[strings.ToLower(sample.Market)] = sample
	}

	return ProductResponse{
		Samples: marketSamples,
	}
}

// header defines the header data.
type header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// extraHeaders defines the extra-headers data for the async call.
type extraHeaders struct {
	Headers []header `json:"headers"`
}

func newExtraHeaders() *extraHeaders {
	return &extraHeaders{
		Headers: make([]header, 0, 100),
	}
}
