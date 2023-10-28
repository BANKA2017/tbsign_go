package _type

type TbsResponse struct {
	Tbs     string `json:"tbs"`
	IsLogin int    `json:"is_login"`
}

type ClientSignResponse struct {
	UserInfo   UserInfo `json:"user_info,omitempty"`
	ContriInfo []any    `json:"contri_info,omitempty"`
	ServerTime string   `json:"server_time,omitempty"`
	Time       int      `json:"time,omitempty"`
	Ctime      int      `json:"ctime,omitempty"`
	Logid      int      `json:"logid,omitempty"`
	ErrorCode  string   `json:"error_code,omitempty"`

	ErrorMsg string `json:"error_msg,omitempty"`
	Info     []any  `json:"info,omitempty"`
}
type AllLevelInfo struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Score string `json:"score,omitempty"`
}
type UserInfo struct {
	UserID           string         `json:"user_id,omitempty"`
	IsSignIn         string         `json:"is_sign_in,omitempty"`
	UserSignRank     string         `json:"user_sign_rank,omitempty"`
	SignTime         string         `json:"sign_time,omitempty"`
	ContSignNum      string         `json:"cont_sign_num,omitempty"`
	TotalSignNum     string         `json:"total_sign_num,omitempty"`
	CoutTotalSingNum string         `json:"cout_total_sing_num,omitempty"`
	HunSignNum       string         `json:"hun_sign_num,omitempty"`
	TotalResignNum   string         `json:"total_resign_num,omitempty"`
	IsOrgName        string         `json:"is_org_name,omitempty"`
	SignBonusPoint   string         `json:"sign_bonus_point,omitempty"`
	MissSignNum      string         `json:"miss_sign_num,omitempty"`
	LevelName        string         `json:"level_name,omitempty"`
	LevelupScore     string         `json:"levelup_score,omitempty"`
	AllLevelInfo     []AllLevelInfo `json:"all_level_info,omitempty"`
	LastLevelScore   string         `json:"last_level_score,omitempty"`
	LastScoreLeft    string         `json:"last_score_left,omitempty"`
	LastLevelName    string         `json:"last_level_name,omitempty"`
	LastLevel        string         `json:"last_level,omitempty"`
}
