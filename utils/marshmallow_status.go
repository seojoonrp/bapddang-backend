// utils/marshmallow_status.go

package utils

func GetMarshmallowStatus(reviewCount int, totalRating int) int {
	if reviewCount <= 0 {
		return -1
	}

	avg := float64(totalRating) / float64(reviewCount)

	switch {
	// [3] 개 노릇노릇한 마시멜로
	// 1. 10끼 이상 & 평균 3.2 이상
	// 2. 7끼 이상 & 평균 4.0 이상
	case (reviewCount >= 10 && avg >= 3.2), (reviewCount >= 7 && avg >= 4.0):
		return 3

	// [2] 나쁘지 않은 마시멜로
	// 1. 10끼 이상 & 평균 2.5 이상
	// 1. 7끼 이상 & 평균 3.2 이상
	// 2. 5끼 이상 & 평균 4.0 이상
	case (reviewCount >= 10 && avg >= 2.5), (reviewCount >= 7 && avg >= 3.2), (reviewCount >= 5 && avg >= 4.0):
		return 2

	// [1] 흰색 마시멜로
	// 1. 7끼 이상 (평균 상관X)
	// 2. 5끼 이상 & 평균 2.5 이상
	// 2. 3끼 이상 & 평균 4.0 이상
	case (reviewCount >= 7), (reviewCount >= 5 && avg >= 2.5), (reviewCount >= 3 && avg >= 4.0):
		return 1

	// [0] 시커멓게 탄 마시멜로
	// - 위 조건에 해당하지 않는 모든 경우 (기록이 1~2개인데 별점도 낮거나, 많이 먹었는데 별점이 엉망인 경우)
	default:
		return 0
	}
}
