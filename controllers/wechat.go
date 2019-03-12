package controllers

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/chanxuehong/wechat/util"

	"github.com/astaxie/beego"
	"github.com/chanxuehong/wechat/mp/core"
	"github.com/chanxuehong/wechat/mp/jssdk"
	"github.com/chanxuehong/wechat/mp/menu"
	"github.com/chanxuehong/wechat/mp/message/callback/request"
	"github.com/chanxuehong/wechat/mp/message/callback/response"
	mpoauth2 "github.com/chanxuehong/wechat/mp/oauth2"
	"github.com/chanxuehong/wechat/oauth2"
)

const (
	wxAppId     = "wx038a0dc5a52a97df"
	wxAppSecret = "cf8a6fc628c00ecedb657b5b97bc4362"

	wxOriId         = "gh_4a3f7b7d54ae"
	wxToken         = "go_jwt"
	wxEncodedAESKey = "fLuzDXZOYq5IauW3J94mX8Tx0B9gfqoRRr4ihfhgk7n"
)

var (
	// 下面两个变量不一定非要作为全局变量, 根据自己的场景来选择.
	msgHandler     core.Handler
	msgServer      *core.Server
	oauth2Endpoint oauth2.Endpoint = mpoauth2.NewEndpoint(wxAppId, wxAppSecret)
)

type WxSignature struct {
	AppID     string `json:"appId"`
	Noncestr  string `json:"noncestr"`
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
	Url       string `json:"url"`
}

type WechatController struct {
	beego.Controller
}

func init() {
	mux := core.NewServeMux()
	mux.DefaultMsgHandleFunc(defaultMsgHandler)
	mux.DefaultEventHandleFunc(defaultEventHandler)
	mux.MsgHandleFunc(request.MsgTypeText, textMsgHandler)
	mux.EventHandleFunc(menu.EventTypeClick, menuClickEventHandler)

	msgHandler = mux
	msgServer = core.NewServer(wxOriId, wxAppId, wxToken, wxEncodedAESKey, msgHandler, nil)
}

func textMsgHandler(ctx *core.Context) {
	log.Printf("收到文本消息:\n%s\n", ctx.MsgPlaintext)

	msg := request.GetText(ctx.MixedMsg)
	resp := response.NewText(msg.FromUserName, msg.ToUserName, msg.CreateTime, msg.Content)
	ctx.RawResponse(resp) // 明文回复
	//ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func defaultMsgHandler(ctx *core.Context) {
	log.Printf("收到消息:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
}

func menuClickEventHandler(ctx *core.Context) {
	log.Printf("收到菜单 click 事件:\n%s\n", ctx.MsgPlaintext)

	event := menu.GetClickEvent(ctx.MixedMsg)
	resp := response.NewText(event.FromUserName, event.ToUserName, event.CreateTime, "收到 click 类型的事件")
	ctx.RawResponse(resp) // 明文回复
	//ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func defaultEventHandler(ctx *core.Context) {
	log.Printf("收到事件:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
}

// wxCallbackHandler 是处理回调请求的 http handler.
func (w *WechatController) WxCallbackHandler() {
	log.Printf("回调处理:\n%s\n", w.Ctx.Request)
	msgServer.ServeHTTP(w.Ctx.ResponseWriter, w.Ctx.Request, nil)
}

// 通过code获取用户openid及用户基本信息
// @router /get_token [post]
func (w *WechatController) GetUserInfo() {
	code := w.GetString("code")

	oauth2Client := oauth2.Client{
		Endpoint: oauth2Endpoint,
	}
	token, err := oauth2Client.ExchangeToken(code)
	if err != nil {
		log.Println(err)
		return
	}

	userinfo, err := mpoauth2.GetUserInfo(token.AccessToken, token.OpenId, "", nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("userinfo: %+v\r\n", userinfo)

	w.Data["json"] = OutResponse(200, userinfo, "")
	w.ServeJSON()
}

// Desc: 自定义分享jsApiticket配置参数
// @router /get_sign [get]
func (w *WechatController) GetSign() {
	var (
		wxSignature       WxSignature
		accessTokenServer core.AccessTokenServer = core.NewDefaultAccessTokenServer(wxAppId, wxAppSecret, nil)
		wechatClient      *core.Client           = core.NewClient(accessTokenServer, nil)
	)

	var ticketServer = jssdk.NewDefaultTicketServer(wechatClient)

	//	fmt.Println(base.GetCallbackIP(wechatClient))

	jsapiTicket, err := ticketServer.Ticket()
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("jsapiTicket: %+v\r\n", jsapiTicket)
	nonceStr := util.NonceStr()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	url := "https://www.unclepang.com"

	signature := jssdk.WXConfigSign(jsapiTicket, nonceStr, timestamp, url)

	wxSignature.AppID = wxAppId
	wxSignature.Noncestr = nonceStr
	wxSignature.Timestamp = timestamp
	wxSignature.Signature = signature
	wxSignature.Url = url

	w.Data["json"] = OutResponse(200, wxSignature, "")
	w.ServeJSON()
}
