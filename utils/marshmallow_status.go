// utils/marshmallow_status.go

package utils

// TODO : 값 조정 필요
func GetMarshmallowStatus(reviewCount int, totalRating int) int {
	averageRating := 0
	if reviewCount > 0 {
		averageRating = totalRating / reviewCount
	}

	switch {
	case reviewCount >= 7 && averageRating >= 4:
		return 3
	case reviewCount >= 5 && averageRating >= 3:
		return 2
	case reviewCount >= 3:
		return 1
	default:
		return 0
	}
}
