package app

type Base struct {
	Id          string `json:"id"`
	Icon        string `json:"icon,omitempty"`        //图标
	Name        string `json:"name"`                  //插件名
	Description string `json:"description,omitempty"` //说明
	Version     string `json:"version,omitempty"`     //版本号 SEMVER v0.0.0
	Internal    bool   `json:"internal,omitempty"`    //内部插件
}

type Menu struct {
	Name       string   `json:"name"`
	Title      string   `json:"title,omitempty"`
	NzIcon     string   `json:"nz_icon,omitempty"` //ant.design图标库
	Items      []*Entry `json:"items,omitempty"`
	Index      int      `json:"index,omitempty"`
	Privileges []string `json:"privileges,omitempty"`
	//Domain     []string `json:"domain"` //域 admin project 或 dealer等
}

type Entry struct {
	Name       string   `json:"name"`
	Title      string   `json:"title,omitempty"`
	Icon       string   `json:"icon,omitempty"`
	Url        string   `json:"url,omitempty"`
	External   bool     `json:"external,omitempty"`
	Privileges []string `json:"privileges,omitempty"`
}

type Privilege struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type App struct {
	Base //继承基础信息

	//资源目录
	Assets string `json:"assets,omitempty"` //资源
	Pages  string `json:"pages,omitempty"`  //页面
	Tables string `json:"tables,omitempty"` //数据表

	//扩展信息
	Type     string `json:"type,omitempty"` //类型
	Author   string `json:"author,omitempty"`
	Email    string `json:"email,omitempty"`
	Homepage string `json:"homepage,omitempty"`

	//资源
	Shortcuts  []*Entry     `json:"shortcuts,omitempty"`  //桌面快捷方式
	Menus      []*Menu      `json:"menus,omitempty"`      //菜单项
	Privileges []*Privilege `json:"privileges,omitempty"` //权限集合

	//前端文件
	Static string `json:"static,omitempty"` //静态目录

	//可执行文件
	Executable   string   `json:"executable,omitempty"` //可执行文件
	Arguments    []string `json:"arguments,omitempty"`  //参数
	Dependencies []string `json:"dependencies,omitempty"`

	//代理
	ApiUrl     string `json:"api_url,omitempty"`
	UnixSocket string `json:"unix_socket,omitempty"`
}
