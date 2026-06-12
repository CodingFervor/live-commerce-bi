package service

import (
	"context"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepo
}

func NewProductService() *ProductService {
	return &ProductService{repo: repository.NewProductRepo()}
}

func (s *ProductService) Create(ctx context.Context, req *model.ProductCreate) (*model.Product, error) {
	p := &model.Product{
		Name:              req.Name,
		Platform:          req.Platform,
		PlatformProductID: req.PlatformProductID,
		Category:          req.Category,
		Brand:             req.Brand,
		Price:             req.Price,
		OriginalPrice:     req.OriginalPrice,
		LivePrice:         req.LivePrice,
		ImageURL:          req.ImageURL,
		Description:       req.Description,
		Tags:              req.Tags,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) List(ctx context.Context, page, pageSize int, platform, category string) ([]model.Product, int, error) {
	return s.repo.List(ctx, page, pageSize, platform, category)
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, id int64, req *model.ProductCreate) (*model.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }
	p.Name = req.Name
	p.Category = req.Category
	p.Brand = req.Brand
	p.Price = req.Price
	p.LivePrice = req.LivePrice
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) GetRankings(ctx context.Context, metric string, limit int, category string) ([]model.ProductRanking, error) {
	if limit <= 0 { limit = 20 }
	return s.repo.GetRankings(ctx, metric, limit, category)
}
