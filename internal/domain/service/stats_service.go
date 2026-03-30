package service

import (
	"context"

	"github.com/go-inventory/internal/domain/model"
)

// About get a inventory time series for a given inventory and window size (number of days)
func (s * WorkerService) GetInventoryTimeSeries(ctx context.Context, windowsize int, offset int, inventory *model.Inventory) (*[]model.Inventory, error){
	result, err := s.callRepositoryRead(ctx, "GetInventoryTimeSeries", func(ctx context.Context) (interface{}, error) {
		return s.workerRepository.GetInventoryTimeSeries(ctx, windowsize, offset, inventory)
	})

	if err != nil {
		return nil, err
	}
	return result.(*[]model.Inventory), nil
}