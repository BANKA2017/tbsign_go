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

type ForumInfo struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	FavoType     string `json:"favo_type,omitempty"`
	LevelID      string `json:"level_id,omitempty"`
	LevelName    string `json:"level_name,omitempty"`
	CurScore     string `json:"cur_score,omitempty"`
	LevelupScore string `json:"levelup_score,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
	Slogan       string `json:"slogan,omitempty"`
}

type ForumListResponse struct {
	ForumList struct {
		NonGconforum *[]ForumInfo `json:"non-gconforum,omitempty"`
		Gconforum    *[]ForumInfo `json:"gconforum,omitempty"`
	} `json:"forum_list,omitempty"`
	HasMore string `json:"has_more,omitempty"`
	// ServerTime string `json:"server_time,omitempty"`
	// Time       int    `json:"time,omitempty"`
	// Ctime      int    `json:"ctime,omitempty"`
	// Logid      int    `json:"logid,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

type WebForumListResponse struct {
	Data struct {
		LikeForum struct {
			List []struct {
				Avatar       string `json:"avatar,omitempty"`
				ForumID      int    `json:"forum_id,omitempty"`
				ForumName    string `json:"forum_name,omitempty"`
				HotNum       int    `json:"hot_num,omitempty"`
				IsBrandForum int    `json:"is_brand_forum,omitempty"`
				LevelID      int    `json:"level_id,omitempty"`
			} `json:"list,omitempty"`
			Page struct {
				CurPage   int `json:"cur_page,omitempty"`
				TotalPage int `json:"total_page,omitempty"`
			} `json:"page,omitempty"`
		} `json:"like_forum,omitempty"`
		// Tokens struct {
		// 	BottomBanner          string `json:"bottom_banner,omitempty"`
		// 	BottomLayer           string `json:"bottom_layer,omitempty"`
		// 	IndexFooterClientDown string `json:"index_footer_client_down,omitempty"`
		// 	IndexMessageIcon      string `json:"index_message_icon,omitempty"`
		// } `json:"tokens,omitempty"`
		// UbsAbtestConfig []struct {
		// 	Sid string `json:"sid,omitempty"`
		// } `json:"ubs_abtest_config,omitempty"`
		// UbsSampleIds string `json:"ubs_sample_ids,omitempty"`
		UserInfo struct {
			ID      int `json:"id,omitempty"`
			IsLogin int `json:"is_login,omitempty"`
		} `json:"user_info,omitempty"`
	} `json:"data,omitempty"`
	Errmsg string `json:"errmsg,omitempty"`
	Errno  int    `json:"errno,omitempty"`
	// Logid      string `json:"logid,omitempty"`
	// ServerTime int    `json:"server_time,omitempty"`
	// Time       string `json:"time,omitempty"`
}

type ForumNameShareResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
	Data  struct {
		Fid         int `json:"fid,omitempty"`
		CanSendPics int `json:"can_send_pics,omitempty"`
	} `json:"data,omitempty"`
	// when fname does not exist in the query string, type of data is string
	//Data string `json:"data,omitempty"`
}

type BaiduUserInfoResponse struct {
	User struct {
		ID       string `json:"id,omitempty"`
		Name     string `json:"name,omitempty"`
		BDUSS    string `json:"BDUSS,omitempty"`
		Portrait string `json:"portrait,omitempty"`
	} `json:"user"`
	Anti struct {
		Tbs string `json:"tbs,omitempty"`
	} `json:"anti,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

type VipInfo struct {
	AScore   int    `json:"a_score,omitempty"`
	ETime    string `json:"e_time,omitempty"`
	ExtScore string `json:"ext_score,omitempty"`
	IconURL  string `json:"icon_url,omitempty"`
	NScore   int    `json:"n_score,omitempty"`
	STime    string `json:"s_time,omitempty"`
	VLevel   int    `json:"v_level,omitempty"`
	VStatus  string `json:"v_status,omitempty"`
	YScore   int    `json:"y_score,omitempty"`
}

type Honor struct {
	Manager struct {
		Assist struct {
			Count     int      `json:"count,omitempty"`
			ForumList []string `json:"forum_list,omitempty"`
		} `json:"assist,omitempty"`
		Manager struct {
			Count     int      `json:"count,omitempty"`
			ForumList []string `json:"forum_list,omitempty"`
		} `json:"manager,omitempty"`
	} `json:"manager,omitempty"`
	Grade map[string]struct {
		Count     int      `json:"count,omitempty"`
		ForumList []string `json:"forum_list,omitempty"`
	} `json:"grade,omitempty"`
	Novice int `json:"novice,omitempty"`
}

type TiebaPanelUserInfoResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
	Data  struct {
		Name string `json:"name,omitempty"`
		// Identity                  any    `json:"identity,omitempty"`
		NameShow                  string `json:"name_show,omitempty"`
		ShowNickname              string `json:"show_nickname,omitempty"`
		ProfessionManagerNickName string `json:"profession_manager_nick_name,omitempty"`
		Portrait                  string `json:"portrait,omitempty"`
		TbAge                     any    `json:"tb_age,omitempty"`
		PostNum                   any    `json:"post_num,omitempty"`
		Honor                     any    `json:"honor,omitempty"`
		VipInfo                   any    `json:"vipInfo,omitempty"`
		TbVip                     bool   `json:"tb_vip,omitempty"`
		FollowedCount             any    `json:"followed_count,omitempty"`
	} `json:"data,omitempty"`
}

type LoginQRCode struct {
	Imgurl string `json:"imgurl,omitempty"`
	Errno  int    `json:"errno,omitempty"`
	Sign   string `json:"sign,omitempty"`
	//Prompt string `json:"prompt,omitempty"`
}

type UnicastResponse struct {
	Errno     int    `json:"errno,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	ChannelV  string `json:"channel_v,omitempty"`
}

type UnicastResponseChannelV struct {
	Status int    `json:"status,omitempty"`
	V      string `json:"v,omitempty"`
	U      any    `json:"u,omitempty"`
}

type WrapUnicastResponse struct {
	ChannelV *UnicastResponseChannelV `json:"channel_v,omitempty"`
	UnicastResponse
}

type LoginResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    struct {
		Session struct {
			Bduss      string `json:"bduss,omitempty"`
			StokenList string `json:"stokenList,omitempty"`
		} `json:"session,omitempty"`
		User struct {
			Username    string `json:"username,omitempty"`
			DisplayName string `json:"displayName,omitempty"`
		} `json:"user,omitempty"`
	} `json:"data,omitempty"`
}

type ManagerTasksResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
	Data  struct {
		BawuTask struct {
			EndTime  int `json:"end_time,omitempty"`
			TaskList []struct {
				TaskName   string `json:"task_name,omitempty"`
				TaskStatus string `json:"task_status,omitempty"`
			} `json:"task_list,omitempty"`
		} `json:"bawu_task,omitempty"`
	} `json:"data,omitempty"`
}
