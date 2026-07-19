package helper

import (
	"context"

	"github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/op"
)

// LLMPriceAddToDB registers model names in the LLM catalog.
// Price fields are left zero: external price sync is disabled for this fork.
func LLMPriceAddToDB(modelNames []string, ctx context.Context) error {
	newLLMInfos := make([]model.LLMInfo, 0, len(modelNames))
	for _, modelName := range modelNames {
		if modelName == "" {
			continue
		}
		newLLMInfos = append(newLLMInfos, model.LLMInfo{Name: modelName})
	}
	if len(newLLMInfos) > 0 {
		return op.LLMBatchCreate(newLLMInfos, ctx)
	}
	return nil
}

// LLMPriceDeleteFromDBWithNoPrice removes catalog rows that still have zero prices.
// Name kept for call-site compatibility with model sync.
func LLMPriceDeleteFromDBWithNoPrice(modelNames []string, ctx context.Context) error {
	if len(modelNames) == 0 {
		return nil
	}
	needDeleteModelNames := make([]string, 0, len(modelNames))
	for _, modelName := range modelNames {
		if modelName == "" {
			continue
		}
		modelPrice, err := op.LLMGet(modelName)
		if err != nil {
			return err
		}
		if modelPrice.Input != 0 || modelPrice.Output != 0 || modelPrice.CacheRead != 0 || modelPrice.CacheWrite != 0 {
			continue
		}
		needDeleteModelNames = append(needDeleteModelNames, modelName)
	}
	if len(needDeleteModelNames) > 0 {
		return op.LLMBatchDelete(needDeleteModelNames, ctx)
	}
	return nil
}
