package hikvideo

type TokenResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AppKey          string `json:"appKey,omitempty"`          //应用唯一标识
		AppAccessToken  string `json:"appAccessToken,omitempty"`  //应用访问凭证
		ExpiresIn       int64  `json:"expiresIn,omitempty"`       //过期时间，单位：小时
		RefreshAppToken string `json:"refreshAppToken,omitempty"` //刷新token
	} `json:"data"`
}

type LiveURLResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		URL string `json:"url"`
	} `json:"data"`
}
