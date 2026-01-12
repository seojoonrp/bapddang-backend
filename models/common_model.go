// models/common_model.go

package models

const (
	FoodTypeStandard = "standard"
	FoodTypeCustom   = "custom"
)

type ReviewFoodItem struct {
	FoodID   string `bson:"food_id" json:"foodID"`
	FoodName string `bson:"food_name" json:"foodName"`
	Type     string `bson:"type" json:"type"`
}
