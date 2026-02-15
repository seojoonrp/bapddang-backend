// utils/marshmallow_status.go

package utils

func GetMarshmallowStatus(reviewCount int, totalRating int) int {
	if reviewCount <= 0 {
		return -1
	}

	avg := float64(totalRating) / float64(reviewCount)

	switch {
	// [3] 개 노릇노릇한 마시멜로
	// 1. 10끼 이상 & 평균 4.3 이상
	// 2. 7끼 이상 & 평균 4.5 이상
	case (reviewCount >= 10 && avg >= 4.3), (reviewCount >= 7 && avg >= 4.5):
		return 3

	// [2] 나쁘지 않은 마시멜로
	// 1. 10끼 이상 & 평균 3.5 이상
	// 1. 7끼 이상 & 평균 3.7 이상
	// 2. 5끼 이상 & 평균 3.9 이상
	case (reviewCount >= 10 && avg >= 3.5), (reviewCount >= 7 && avg >= 3.7), (reviewCount >= 5 && avg >= 3.9):
		return 2

	// [0] 시커멓게 탄 마시멜로
	// 1. 10끼 이상 & 평균 3.5 미만
	// 2. 7끼 이상 & 평균 3.7 미만
	// 2. 5끼 이상 & 평균 3.3 미만
	case (reviewCount >= 10 && avg < 3.5), (reviewCount >= 7 && avg < 3.7), (reviewCount >= 5 && avg < 3.3):
		return 0

	// [1]
	// - 위 조건에 해당하지 않는 모든 경우
	default:
		return 1
	}
}
