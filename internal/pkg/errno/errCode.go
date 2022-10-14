package errno

type Erron struct {
	Code    int
	Message string
}

var (
	//web
	PushNoName         = &Erron{Code: 1000, Message: "未提供索引名称"}       // push未携带参数
	PushNoFile         = &Erron{Code: 1001, Message: "任务名称未找到对应配置文件"} //config.json没配置
	PushGetIndexStatus = &Erron{Code: 1002}                           //push 控制器中 GetIndexStatus抛出的错误
	PushIndexExist     = &Erron{Code: 1003, Message: "Index名称已经存在"}   // 存在Index 就没法用 Index_A Index_B 加别名的方式区分
	PushCreateIndex    = &Erron{Code: 1004}                           //首次创建索引失败
	PushError          = &Erron{Code: 1005}                           //正式推送中的错误
	//sys
	SysConfigNotFind                   = "配置文件未找到"
	SysAliasExceedLimit                = "别名下存在多个Index"
	SysIndexAliasExceedLimit           = "Index下包含多个别名"
	SysIndexNameStandardLimit          = "Index命名非下划线A或者B"
	SysIndexGetInfoTransitionJsonError = "Index详细信息格式转化失败"
	SysIndexExist                      = "Index名已经存在"
	SysTypeUndefined                   = "column undefined"
)
