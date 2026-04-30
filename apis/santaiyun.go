package apis

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
)

const login2URL = "http://www.santaiyun.com/Admin/Api/loginCheck/username/%s/pass/%s"

type login2Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func init() {
	api.RegisterUnAuthorized("GET", "login2", login2)
}

func login2(ctx *gin.Context) {
	username := ctx.Query("username")
	pass := ctx.Query("pass")

	if username == "" || pass == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "username和pass不能为空",
			"data":    []any{},
		})
		return
	}
	//
	//resp, err := http.PostForm(login2URL, url.Values{
	//	"username": {username},
	//	"pass":     {pass},
	//})
	u := fmt.Sprintf(login2URL, username, pass)
	resp, err := http.Get(u)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "请求三泰云登录接口失败: " + err.Error(),
			"data":    []any{},
		})
		return
	}
	defer resp.Body.Close()

	var result login2Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "解析三泰云返回结果失败: " + err.Error(),
			"data":    []any{},
		})
		return
	}

	if len(result.Data) == 0 {
		result.Data = json.RawMessage("[]")
	}

	if result.Code == 1 {
		session := sessions.Default(ctx)
		session.Set("user", username)
		session.Set("admin", true)
		session.Set("tenant", "")
		if err := session.Save(); err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "登录成功，但保存session失败: " + err.Error(),
				"data":    []any{},
			})
			return
		}
	}

	//ctx.JSON(http.StatusOK, result)
	ctx.Redirect(http.StatusTemporaryRedirect, "/")
}
