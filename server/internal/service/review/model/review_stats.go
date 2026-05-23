package model

// ReviewStats 评价统计
type ReviewStats struct {
	TotalCount    int64   `json:"total_count"`    //总评价数
	Rating5Count  int64   `json:"rating_5_count"` //5星评价数
	Rating4Count  int64   `json:"rating_4_count"`
	Rating3Count  int64   `json:"rating_3_count"`
	Rating2Count  int64   `json:"rating_2_count"`
	Rating1Count  int64   `json:"rating_1_count"`
	AverageRating float64 `json:"average_rating"` //平均评分
}
