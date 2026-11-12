package product

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryRepo "autostack/internal/repository/inventory"
	productRepo "autostack/internal/repository/product"
	"github.com/gin-gonic/gin"
)

type failingProductRepository struct {
	productRepo.ProductRepository
	reads     int
	reloadErr error
}

func (r *failingProductRepository) FindByID(context.Context, uint) (*Product, error) {
	r.reads++
	if r.reads > 1 {
		return nil, r.reloadErr
	}
	return &Product{ID: 1}, nil
}
func (r *failingProductRepository) UpdateFields(context.Context, uint, map[string]interface{}) error {
	return nil
}
func (r *failingProductRepository) List(context.Context, *productRepo.ProductQuery) ([]Product, int64, error) {
	return []Product{{ID: 1}}, 1, nil
}

type failingWarehouseRepository struct {
	inventoryRepo.WarehouseRepository
	err error
}

func (r failingWarehouseRepository) List(context.Context, *inventoryRepo.WarehouseQuery) ([]Warehouse, error) {
	return nil, r.err
}

type failingSupplierRepository struct {
	productRepo.ProductSupplierRepository
	err error
}

func (r failingSupplierRepository) FindDefaultByProductIDs(context.Context, []uint) ([]ProductSupplier, error) {
	return nil, r.err
}

func TestUpdateProductPropagatesReloadFailure(t *testing.T) {
	want := errors.New("reload failed")
	s := &Service{productRepo: &failingProductRepository{reloadErr: want}}
	product, err := s.UpdateProduct(90000000001, 1, UpdateProductRequest{})
	if !errors.Is(err, want) || product != nil {
		t.Fatalf("got product=%v err=%v", product, err)
	}
}

func TestListProductsWithSupplierPropagatesSupplementalQueryErrors(t *testing.T) {
	for _, stage := range []string{"warehouse", "supplier"} {
		t.Run(stage, func(t *testing.T) {
			want := errors.New(stage + " query failed")
			warehouse := failingWarehouseRepository{}
			supplier := failingSupplierRepository{}
			if stage == "warehouse" {
				warehouse.err = want
			} else {
				supplier.err = want
			}
			s := &Service{productRepo: &failingProductRepository{}, warehouseRepo: warehouse, supplierRepo: supplier}
			rows, _, err := s.ListProductsWithSupplier(90000000001, 1, 20, "", 0)
			if !errors.Is(err, want) || rows != nil {
				t.Fatalf("got rows=%v err=%v", rows, err)
			}
		})
	}
}

func TestListProductsReturnsErrorWhenWarehouseLookupFails(t *testing.T) {
	previous := service
	t.Cleanup(func() { service = previous })
	service = &Service{productRepo: &failingProductRepository{}, warehouseRepo: failingWarehouseRepository{err: errors.New("warehouse query failed")}}
	router := gin.New()
	router.GET("/products", func(c *gin.Context) { c.Set("company_id", uint(90000000001)); ListProducts(c) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/products", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
