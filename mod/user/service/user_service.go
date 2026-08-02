package service

import (
	context2 "context"
	"strings"
	"time"

	"github.com/example/go-ai-scaffold/mod/user/dao/departmentdao"
	"github.com/example/go-ai-scaffold/mod/user/dao/roledao"
	"github.com/example/go-ai-scaffold/mod/user/dao/userdao"
	"github.com/example/go-ai-scaffold/mod/user/model"
	"github.com/example/go-ai-scaffold/pkg/class"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/library/cryptokit"
	"github.com/example/go-ai-scaffold/pkg/library/stringkit"
	"github.com/example/go-ai-scaffold/pkg/service/rediskit"
	"github.com/example/go-ai-scaffold/pkg/service/sqlkit"
)

// ============ Login ============

// Login 合并 loginByUsername 和 login，返回用户。
// 调用方负责创建 JWT、设置 cookie、调用 AdditionLoginFunc 等 HTTP 层逻辑。
func Login(username, phone, pwd string) *model.User {
	if stringkit.IsNull(username) && stringkit.IsNull(phone) {
		panic(exception.New("用户名或手机号缺失"))
	}
	username = strings.TrimSpace(username)
	phone = strings.TrimSpace(phone)
	pwd = cryptokit.MD5(pwd)
	dao := userdao.New(userdao.OptsDefault)
	user := dao.Login(pwd, username, phone)
	if user == nil {
		panic(exception.New("账号和密码不匹配"))
	}
	if user.Status.Int32 == model.UserStatusFreeze {
		panic(exception.New("账户被冻结"))
	}
	return user
}

// ============ User Info ============

// GetUserById 根据 uid 获取用户。
func GetUserById(uid int64) *model.User {
	dao := userdao.New(userdao.OptsDefault)
	return dao.SelectOneById(uid)
}

// ============ UpdatePwd ============

// UpdatePwd 修改密码。
func UpdatePwd(uid int64, oldPwd, newPwd string) {
	dao := userdao.New(userdao.OptsDefault)
	user := dao.SelectOneById(uid)
	if user == nil { // B1: nil check
		panic(exception.New("用户不存在"))
	}
	if user.Pwd.String != cryptokit.MD5(oldPwd) {
		panic(exception.New("原密码错误"))
	}
	user.Pwd.Set(cryptokit.MD5(newPwd))
	dao.UpdateObj(user)
}

// ============ UpdateUserInfo ============

type UpdateUserInfoParams struct {
	Username   class.String
	Name       class.String
	Phone      class.String
	Sms        class.String
	Gender     int8
	Image      class.String
	Address    class.String
	OldPwd     class.String
	NewPwd     class.String
	ExtendJson class.MapString
}

// UpdateUserInfo 用户自助修改信息。
// B3: 用 SelectOneById 而非 SelectOneWithDelById，不允许修改已删除用户。
// B4: 改密码分支增加 nil check。
func UpdateUserInfo(uid int64, params UpdateUserInfoParams) {
	dao := userdao.New(userdao.OptsNone)
	u := dao.SelectOneById(uid)
	if u == nil { // B1: nil check
		panic(exception.New("用户不存在"))
	}
	if params.Phone.Valid && params.Phone.String != "" && params.Phone.String != u.Phone.String {
		if dao.FindByPhone(params.Phone.String) != nil {
			panic(exception.New("手机号已被注册"))
		}
		if !params.Sms.Valid || rediskit.Get(context2.Background(), rediskit.GetKeyWithPrefix("sms:"+params.Phone.String), "") != params.Sms.String {
			panic(exception.New("验证码错误"))
		}
	}
	if params.Username.Valid && params.Username.String != u.Username.String {
		if dao.FindByUsername(params.Username.String) != nil {
			panic(exception.New("该用户名已被使用"))
		}
		u.Username.Set(params.Username.String)
	}
	if params.Image.Valid {
		u.Image.Set(params.Image)
	}
	if params.Name.Valid {
		u.Name.Set(params.Name.String)
	}
	if params.Phone.Valid {
		u.Phone.Set(params.Phone.String)
	}
	if params.Gender != 0 {
		u.Gender.Set(params.Gender)
	}
	if params.Address.Valid {
		u.Address.Set(params.Address.String)
	}
	if params.ExtendJson.Valid {
		u.Extend.PutAll(params.ExtendJson.Map)
	}
	if params.OldPwd.Valid && params.NewPwd.Valid && params.OldPwd.String != "" && params.NewPwd.String != "" {
		user := dao.SelectOneById(u.Id)
		if user == nil { // B4: nil check
			panic(exception.New("用户不存在"))
		}
		if user.Pwd.String != cryptokit.MD5(params.OldPwd.String) {
			panic(exception.New("原密码错误"))
		}
		user.Pwd.Set(cryptokit.MD5(params.NewPwd.String))
	}
	dao.UpdateObj(u)
}

// ============ ListUsers ============

type ListUsersParams struct {
	DepartmentIds []int64
	RoleIds       []int64
}

// ListUsers 管理员查看用户列表。
func ListUsers(params ListUsersParams) []*model.User {
	dao := userdao.New(userdao.OptsDefault)
	return dao.List(userdao.ListParam{Roles: params.RoleIds, Departments: params.DepartmentIds})
}

// ============ AddUser ============

type AddUserParams struct {
	Username   class.String `validate:"required"`
	Pwd        class.String `validate:"required"`
	Role       class.Int64
	Department class.Int64
	Name       class.String
	Phone      class.String
	Sms        class.String
	Gender     int8
	Image      class.String
	Address    class.String
	ExtendJson class.MapString
}

// AddUser B7: 不复用已删除用户记录，避免脏数据。B8: 查 role 用 OptsNone。
func AddUser(params AddUserParams, checkSms bool) *model.User {
	dao := userdao.New(userdao.OptsNone)
	if dao.FindByUsername(params.Username.String) != nil {
		panic(exception.New("用户名已经存在"))
	}
	if params.Phone.Valid && dao.FindByPhone(params.Phone.String) != nil {
		panic(exception.New("手机号已经存在"))
	}
	if params.Phone.Valid && checkSms && (!params.Sms.Valid || rediskit.Get(context2.Background(), rediskit.GetKeyWithPrefix("sms:"+params.Phone.String), "") != params.Sms.String) {
		panic(exception.New("验证码错误"))
	}
	u := &model.User{}
	u.CreateDt.Set(time.Now())
	if params.Role.IsValid() {
		// B8: 只需 role 自身，无需级联 department
		roleDao := roledao.New(roledao.OptsNone)
		r := roleDao.SelectOneById(params.Role)
		if r == nil {
			panic(exception.New("角色不存在"))
		}
		u.Role = r
		if !params.Department.IsValid() && r.Department != nil {
			u.Department = r.Department
		}
	}
	if params.Department.IsValid() {
		deptDao := departmentdao.New(departmentdao.OptsNone)
		dept := deptDao.SelectOneById(params.Department.Int64)
		if dept == nil {
			panic(exception.New("部门不存在"))
		}
		u.Department = dept
	}
	if params.Username.Valid {
		u.Username.Set(params.Username)
	}
	if params.Pwd.Valid {
		u.Pwd.Set(cryptokit.MD5(params.Pwd.String))
	}
	if params.Name.Valid {
		u.Name.Set(params.Name)
	}
	if params.Phone.Valid {
		u.Phone.Set(params.Phone)
	}
	if params.Image.Valid {
		u.Image.Set(params.Image)
	}
	if params.Gender != 0 {
		u.Gender.Set(params.Gender)
	}
	if params.Address.Valid {
		u.Address.Set(params.Address)
	}
	if params.ExtendJson.Valid {
		u.Extend.PutAll(params.ExtendJson.Map)
	}
	dao.InsertObj(u)
	return u
}

// ============ UpdateUser ============

type UpdateUserParams struct {
	Id         int64 `validate:"required"`
	Username   class.String
	Name       class.String
	Phone      class.String
	Gender     int8
	Image      class.String
	Address    class.String
	Pwd        class.String
	Role       class.Int64
	Department class.Int64
	ExtendJson class.MapString
}

// UpdateUser 管理员修改用户。
func UpdateUser(params UpdateUserParams) {
	dao := userdao.New(userdao.OptsDefault)
	u := dao.SelectOneById(params.Id)
	if u == nil {
		panic(exception.New("用户不存在"))
	}
	// B9: Role.Id == 0 表示该用户绑定了内置超级管理员角色，不允许修改
	if u.Role != nil && u.Role.Id == 0 {
		panic(exception.New("该用户不能设置"))
	}
	if params.Phone.Valid && params.Phone.String != "" && params.Phone.String != u.Phone.String && dao.FindByPhone(params.Phone.String) != nil {
		panic(exception.New("手机号已存在"))
	}
	if params.Username.Valid && params.Username.String != u.Username.String {
		if dao.FindByUsername(params.Username.String) != nil {
			panic(exception.New("该用户名已被使用"))
		}
		u.Username.Set(params.Username.String)
	}
	if params.Role.Int64 > 0 && (u.Role == nil || params.Role.Int64 != u.Role.Id) {
		rdao := roledao.New(roledao.OptsDefault)
		r := rdao.SelectOneById(params.Role)
		if r == nil {
			panic(exception.New("role不存在"))
		}
		u.Role = r
		if !params.Department.IsValid() {
			u.Department = r.Department
		}
	}
	if params.Department.IsValid() && (u.Department == nil || params.Department.Int64 != u.Department.Id) {
		deptDao := departmentdao.New(departmentdao.OptsNone)
		dept := deptDao.SelectOneById(params.Department.Int64)
		if dept == nil {
			panic(exception.New("部门不存在"))
		}
		u.Department = dept
	}
	if params.Name.Valid {
		u.Name.Set(params.Name.String)
	}
	if params.Phone.Valid {
		u.Phone.Set(params.Phone.String)
	}
	if params.Image.Valid {
		u.Image.Set(params.Image)
	}
	if params.Pwd.Valid && params.Pwd.String != "" {
		u.Pwd.Set(cryptokit.MD5(params.Pwd.String))
	}
	if params.Gender != 0 {
		u.Gender.Set(params.Gender)
	}
	if params.Address.Valid {
		u.Address.Set(params.Address.String)
	}
	if params.ExtendJson.Valid {
		u.Extend.PutAll(params.ExtendJson.Map)
	}
	dao.UpdateObj(u)
}

// ============ DeleteUser ============

type DeleteUserParams struct {
	Id  int64       `validate:"required"`
	Off class.Int32 `validate:"required" comment:"0-删除，1-冻结，2-解冻"`
}

// DeleteUser B10: 用 OptsNone 避免级联查询浪费。
func DeleteUser(operatorUid int64, params DeleteUserParams) {
	if operatorUid == 0 {
		panic(exception.New("登录的用户错误"))
	}
	if operatorUid == params.Id {
		panic(exception.New("不能操作自己"))
	}
	dao := userdao.New(userdao.OptsNone) // B10: OptsNone
	target := dao.SelectOneById(params.Id)
	if target == nil {
		panic(exception.New("用户不存在"))
	}
	// B9: Role.Id == 0 表示内置超级管理员，不允许删除/冻结
	if target.Role != nil && target.Role.Id == 0 {
		panic(exception.New("该用户不能设置"))
	}
	if target.Extend.GetBool("immutable") {
		panic(exception.New("该用户不可删除"))
	}
	sqlkit.TxArea(func(targetDS *sqlkit.DataSource) {
		dao1 := userdao.New(userdao.OptsNone, targetDS) // B10: OptsNone
		switch params.Off.Int32 {
		case 0:
			dao1.SetNull(params.Id)
			dao1.DeleteById(params.Id)
		case 1:
			dao1.FreezeUser(params.Id, model.UserStatusFreeze)
		case 2:
			dao1.FreezeUser(params.Id, model.UserStatusOK)
		default:
			panic(exception.New("无效的操作类型"))
		}
	})
}
