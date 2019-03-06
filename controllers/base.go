package controllers

import (
	_const "go_wechat/const"
	"time"

	"github.com/astaxie/beego"
)

const TimeFormart = "2006-01-02 15:04:05"

// Operations about Main
type MainController struct {
	beego.Controller
}

//map[string]interface{}{"code": 400, "msg": "no user exists!", "data": nil}
type Response struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	TimeStamp int64       `json:"timestamp"`
}

func OutResponse(code int, data interface{}, msg string) Response {
	if msg == "" {
		msg = _const.StatusText(code)
	}
	resp := Response{
		Code:      code,
		Msg:       msg,
		Data:      data,
		TimeStamp: time.Now().Unix(), //time.Now().Format("2006-01-02 15:04:05")
	}
	return resp
}

func (m *MainController) Welcome() {
	m.Data["Website"] = "www.unclepang.com"
	m.Data["Email"] = "10846295@qq.com"
	m.TplName = "index.tpl"
	m.Render()
}
