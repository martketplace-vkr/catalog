package dto

import (
	"strings"

	adminpb "github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/admin"
)

type CreateCategoryRequest struct {
	Name     string
	ParentID *int64
}

type UpdateCategoryRequest struct {
	CategoryID int64
	Name       string
	ParentID   *int64
}

type UpdateExchangeRateRequest struct {
	RubPerUSDT string
}

func CreateCategoryRequestFromProto(req *adminpb.CreateCategoryRequest) CreateCategoryRequest {
	if req == nil {
		return CreateCategoryRequest{}
	}

	return CreateCategoryRequest{
		Name:     strings.TrimSpace(req.Name),
		ParentID: req.ParentId,
	}
}

func UpdateCategoryRequestFromProto(req *adminpb.UpdateCategoryRequest) UpdateCategoryRequest {
	if req == nil {
		return UpdateCategoryRequest{}
	}

	return UpdateCategoryRequest{
		CategoryID: req.CategoryId,
		Name:       strings.TrimSpace(req.Name),
		ParentID:   req.ParentId,
	}
}

func UpdateExchangeRateRequestFromProto(req *adminpb.UpdateUSDTExchangeRateRequest) UpdateExchangeRateRequest {
	if req == nil {
		return UpdateExchangeRateRequest{}
	}

	return UpdateExchangeRateRequest{
		RubPerUSDT: strings.TrimSpace(req.RubPerUsdt),
	}
}
