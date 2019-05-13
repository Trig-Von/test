package routers

import (
	"github.com/astaxie/beego/context"
	"pyg/pyg/controllers"
	"github.com/astaxie/beego"
)

func init() {
	//路由过滤器
	beego.InsertFilter("/user/*",beego.BeforeExec,filtFunc)
	beego.Router("/", &controllers.MainController{})
	//用户注册
	beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
	//发送短信
	beego.Router("/sendMsg",&controllers.UserController{},"post:HandleSendMsg")
	//邮箱激活
	beego.Router("/register-email",&controllers.UserController{},"get:ShowEmail;post:HandleEmail")
	//激活用户
	beego.Router("/active",&controllers.UserController{},"get:Active")
	//登录用户
	beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")
	//展示首页
	beego.Router("/index",&controllers.GoodsController{},"get:ShowIndex")
	//推出登录
	beego.Router("/user/logout",&controllers.UserController{},"get:LogOut")
	//展示用户中心
	beego.Router("/user/userCenterInfo",&controllers.UserController{},"get:ShowUserCenterInfo")
	//收获地址页
	beego.Router("/user/site",&controllers.UserController{},"get:ShowSite;post:HandleSite")
}

func filtFunc(ctx *context.Context)  {
	userName := ctx.Input.Session("userName")
	if userName == nil {
		ctx.Redirect(302,"/login")
		return
	}
}