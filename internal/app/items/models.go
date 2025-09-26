package items

type CreateItemRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=30"`
	ProductCategory string `json:"productCategory" validate:"required,oneof=Beverage Food Snack Condiments Additions"`
	Price           int64  `json:"price" validate:"required,min=1"` // dalam satuan terkecil (misal: rupiah tanpa koma)
	ImageUrl        string `json:"imageUrl" validate:"required,url"`
}
