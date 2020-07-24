package remote

type ResponseInfo struct {
	ErrCode string      `json:"err_code"` // 错误码
	ErrCue  string      `json:"err_cue"`  // 错误提示语
	ErrMsg  string      `json:"err_msg"`  // 错误原因
	Data    interface{} `json:"data"`
}
