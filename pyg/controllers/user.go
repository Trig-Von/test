package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"math/rand"
	"pyg/pyg/models"
	"regexp"
	"time"
)

type UserController struct {
	beego.Controller
}


func(this*UserController)ShowRegister(){
	this.TplName = "register.html"
}

func RespFunc(this* UserController,resp map[string]interface{}){
	//3.把容器传递给前段
	this.Data["json"] = resp
	//4.指定传递方式
	this.ServeJSON()
}

type Message struct {
	Message string
	RequestId string
	BizId string
	Code string
}

//发送短信
func(this*UserController)HandleSendMsg(){
	//接受数据
	phone := this.GetString("phone")
	resp := make(map[string]interface{})

	defer RespFunc(this,resp)
	//返回json格式数据
	//校验数据
	if phone == ""{
		beego.Error("获取电话号码失败")
		//2.给容器赋值
		resp["errno"] = 1
		resp["errmsg"] = "获取电话号码错误"
		return
	}
	//检查电话号码格式是否正确
	reg,_ :=regexp.Compile(`^1[3-9][0-9]{9}$`)
	result := reg.FindString(phone)
	if result == ""{
		beego.Error("电话号码格式错误1")
		//2.给容器赋值
		resp["errno"] = 2
		resp["errmsg"] = "电话号码格式错误"
		return
	}
	//发送短信   SDK调用
	client, err := sdk.NewClientWithAccessKey("cn-hangzhou", "LTAIu4sh9mfgqjjr", "sTPSi0Ybj0oFyqDTjQyQNqdq9I9akE")
	if err != nil {
		beego.Error("电话号码格式错误2")
		//2.给容器赋值
		resp["errno"] = 3
		resp["errmsg"] = "初始化短信错误"
		return
	}
	//生成6位数随机数
	rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode :=fmt.Sprintf("%06d",rand.Int31n(1000000))

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = "cn-hangzhou"
	request.QueryParams["PhoneNumbers"] = phone
	request.QueryParams["SignName"] = "品优购"
	request.QueryParams["TemplateCode"] = "SMS_164275022"
	request.QueryParams["TemplateParam"] = "{\"code\":"+vcode+"}"

	response, err := client.ProcessCommonRequest(request)
	if err != nil {

		beego.Error("电话号码格式错误3",err)
		//2.给容器赋值
		resp["errno"] = 4
		resp["errmsg"] = "短信发送失败"
		return
	}
	//json数据解析
	var message Message
	json.Unmarshal(response.GetHttpContentBytes(),&message)
	if message.Message != "OK"{
		beego.Error("电话号码格式错误4")
		//2.给容器赋值
		resp["errno"] = 6
		resp["errmsg"] = message.Message
		return
	}

	resp["errno"] = 5
	resp["errmsg"] = "发送成功"
	resp["code"] = vcode
}

func (this *UserController)HandleRegister()  {
	//获取数据
	phone := this.GetString("phone")
	pwd := this.GetString("password")
	rpwd := this.GetString("repassword")
	//校验数据
	if phone == "" || pwd == "" || rpwd == "" {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "获取数据错误"
		this.TplName = "register.html"
		return
	}
	if pwd != rpwd {
		beego.Error("两次输入密码不一直")
		this.Data["errmsg"] = "两次输入密码不一直"
		this.TplName = "register.html"
		return
	}
	//处理数据
	//orm插入数据
	o := orm.NewOrm()
	var user models.User
	user.Name = phone
	user.Pwd = pwd
	user.Phone = phone
	o.Insert(&user)
	//激活页面
	this.Ctx.SetCookie("userName",user.Name,600)
	this.Redirect("/register-email",302)
	//返回数据
}

//展示邮箱激活
func (this *UserController)ShowEmail()  {
	this.TplName = "register-email.html"
}

//处理邮箱激活业务
func (this *UserController)HandleEmail()  {
	//获取数据
	email := this.GetString("email")
	pwd := this.GetString("password")
	rpwd := this.GetString("repassword")
	//校验数据
	if email == "" || pwd == "" || rpwd == "" {
		beego.Error("输入数据不完整")
		this.Data["errmsg"] = "输入数据不完整"
		this.TplName = "register-email.html"
		return
	}
	//两次密码是否一致
	if pwd != rpwd {
		beego.Error("两次密码输入不一致")
		this.Data["errmsg"] = "两次密码输入不一致"
		this.TplName = "register-email.html"
		return
	}
	//校验邮箱格式
	reg,_ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result := reg.FindString(email)
	if result == "" {
		beego.Error("邮箱格式错误")
		this.Data["errmsg"] = "邮箱格式错误"
		this.TplName = "register-email.html"
		return
	}

	//处理数据
	//发送邮件
	//utils  全局通用接口  工具类    邮箱配置
	config := `{"username":"1510271838@qq.com","password":"ynojniemjvbnigch","host":"smtp.qq.com","port":587}`
	emailReg := utils.NewEMail(config)
	//内容配置
	emailReg.Subject = "品邮购用户激活"
	emailReg.From = "1510271838@qq.com"
	emailReg.To = []string{email}
	userName := this.Ctx.GetCookie("userName")
	emailReg.HTML = `<a href="http://192.168.230.81:8080/active?userName=`+userName+`"> 点击激活该用户</a>`

	//发送
	emailReg.Send()

	//返回数据
	this.Ctx.WriteString("邮件已发送，请前往邮箱激活目标")
}

//激活
func (this *UserController)Active()  {
	//获取数据
	userName := this.GetString("userName")
	//校验数据
	if userName == "" {
		beego.Error("用户名错误")
		this.Redirect("/register-email",302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user,"Name")
	if err != nil {
		beego.Error("用户名不存在")
		this.Redirect("/register-email",302)
		return
	}
	user.Active = true
	o.Update(&user,"Active")

	//返回数据
	this.Redirect("/login",302)
}

//展示登录页面
func (this *UserController)ShowLogin()  {
	userName := this.Ctx.GetCookie("userName")
	//解密
	dec,_ := base64.StdEncoding.DecodeString(userName)
	if userName != "" {
		this.Data["userName"] = string(dec)
		this.Data["checked"] = "checked"
	}else {
		this.Data["userName"] = ""
		this.Data["checked"] = ""
	}

	this.TplName = "login.html"
}

//处理用户登录
func (this *UserController)HandleLogin()  {
	//获取数据
	userName := this.GetString("userName")
	pwd := this.GetString("password")
	//校验数据
	if userName == "" || pwd == "" {
		beego.Error("传输数据不完整")
		this.TplName = "login.html"
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.User
	reg,_ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result := reg.FindString(userName)
	if result != "" {
		user.Email = userName
		err := o.Read(&user,"Email")
		if err != nil {
			beego.Error("用户名不存在")
			this.TplName = "login.html"
			return
		}

		if user.Pwd!= pwd{
			beego.Error("密码错误")
			this.TplName = "login.html"
			return
		}
	}else {
		user.Name = userName
		err := o.Read(&user,"Name")
		if err != nil {
			beego.Error("用户名不存在")
			this.TplName = "login.html"
			return
		}
		if user.Pwd != pwd {
			beego.Error("密码错误")
			this.TplName = "login.html"
			return
		}
	}

	if user.Active == false  {
		this.Data["errmsg"] = "未激活，请激活"
		this.TplName = "login.html"
		return
	}


	//实现记住用户名功能  上一次登陆成功以后，点击了记住用户名，下一次登陆的时候默认显示用户名
	remember:= this.GetString("m1")
	//给userName加密
	enc := base64.StdEncoding.EncodeToString([]byte(userName))
	if remember == "2"{
		this.Ctx.SetCookie("userName",enc,60*60)
	}else {
		this.Ctx.SetCookie("userName",userName,-1)
	}

	//session存储
	this.SetSession("userName",userName)
	//返回数据
	this.Redirect("/index",302)
}

func (this *UserController)LogOut()  {
	this.DelSession("userName")
	this.Redirect("/login",302)
}

//展示用户中心页
func (this *UserController)ShowUserCenterInfo()  {
	this.Data["change"] = "1"
	this.Layout = "layout.html"
	this.TplName = "user_center_info.html"
}

//展示用户中心地址页
func (this *UserController)ShowSite()  {

	o := orm.NewOrm()
	var address models.Address
	name := this.GetSession("userName")
	qs := o.QueryTable("Address").RelatedSel("User").Filter("User__Name",name.(string))
	qs.Filter("IsDefault",true).One(&address)

	this.Data["address"] = address

	this.Data["change"] = "3"
	this.Layout = "layout.html"
	this.TplName = "user_center_site.html"
}

func (this *UserController)HandleSite()  {
	//获取数据
	receiver := this.GetString("receiver")
	addrdetail := this.GetString("addrdetail")
	postCode := this.GetString("postCode")
	phone := this.GetString("phone")
	//校验数据
	if receiver == "" || addrdetail == "" || postCode == "" || phone == "" {
		beego.Error("获取数据错误")
		this.Layout = "layout.html"
		this.TplName = "user_center_site.html"
		return
	}

	//处理数据
	o := orm.NewOrm()
	var userAddr models.Address
	userAddr.Receiver = receiver
	userAddr.Addr = addrdetail
	userAddr.PostCode = postCode
	userAddr.Phone = phone

	name := this.GetSession("userName")
	var user models.User
	user.Name = name.(string)
	o.Read(&user,"Name")
	userAddr.User = &user

	var oldAddress models.Address
	qs := o.QueryTable("Address").RelatedSel("User").Filter("User__Name",name.(string))
	err:= qs.Filter("IsDefault",true).One(&oldAddress)

	if err == nil {
		oldAddress.IsDefault = false
		o.Update(&oldAddress,"IsDefault")
	}
	userAddr.IsDefault = true

	_,err = o.Insert(&userAddr)
	if err != nil {
		beego.Error("插入失败",err)
		this.Layout = "layout.html"
		this.TplName = "user_center_site.html"
		return
	}
	//返回数据
	this.Redirect("/user/site",302)
}