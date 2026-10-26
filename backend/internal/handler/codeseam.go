package handler

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (h *GatewayHandler) CodeSeamResource(ctx context.Context, k *service.APIKey) (*service.CodeSeamResource, error) {
	if h == nil || h.gatewayService == nil {
		return nil, nil
	}
	return h.gatewayService.CodeSeamResource(ctx, k)
}
