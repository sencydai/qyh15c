package protocol

const (
	Chat   byte = 9   //聊天
	Fight  byte = 52  //战斗
	Fuben  byte = 181 // 副本
	System byte = 255 // 系统
)

//聊天
const (
	ChatCSendChatMsg = 1
)

// 战斗
const (
	FightCGetAwards byte = 2 // 领取奖励

	FightSResult byte = 1 //战斗结果
)

// 副本
const (
	FubenCLoginMainFuben byte = 2 //登陆主线副本
)

// 系统
const (
	SystemCLogin       byte = 1 // 登陆
	SystemCCreateActor byte = 2 // 创建角色
	SystemCActorList   byte = 4 // 查询角色列表
	SystemCLoginGame   byte = 5 //登陆游戏
	SystemCRandomName  byte = 6 // 随机名称

	SystemSLogin       byte = 1 //登陆
	SystemSCreateActor byte = 2 // 创建角色
	SystemSActorLists  byte = 4 //查询角色列表
	SystemSLoginGame   byte = 5 //登陆游戏
	SystemSRandomName  byte = 6 //随机名称
)
