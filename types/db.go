package _type

type StatusStruct struct {
	Success  int64 `json:"success"`
	Failed   int64 `json:"failed"`
	Waiting  int64 `json:"waiting"`
	IsIgnore int64 `json:"ignore"` // `ignore` is the keyword of SQLite
}
