package main

import (
	"bytes"

	"xgame/proto/pack"
	proto "xgame/proto/protocol"

	"github.com/sencydai/gameworld/base"
)

type LordData struct {
	decor map[int]*LordDecorData
	equip map[int]*LordEquipData
}

type LordDecorData struct {
	id     int
	unLock map[int]bool
}

type LordEquipData struct {
	stage int
	level int
}

func GetLordData(account *Account) *LordData {
	if data, ok := account.data[keyLord]; ok {
		return data.(*LordData)
	}

	account.data[keyLord] = &LordData{}
	return account.data[keyLord].(*LordData)
}

func GetLordDecors(account *Account) map[int]*LordDecorData {
	lordData := GetLordData(account)
	if lordData.decor == nil {
		lordData.decor = make(map[int]*LordDecorData)
	}
	return lordData.decor
}

func GetLordEquips(account *Account) map[int]*LordEquipData {
	lordData := GetLordData(account)
	if lordData.equip == nil {
		lordData.equip = make(map[int]*LordEquipData)
	}
	return lordData.equip
}

func init() {
	//领主装饰初始化
	RegServerHandle(proto.Lord, proto.LordSDecorInit, HandleLordDecorInit)
	//装饰解锁
	RegServerHandle(proto.Lord, proto.LordSDecorUnlock, HandleLordDecorUnlock)
	//装备初始化
	RegServerHandle(proto.Lord, proto.LordSEquipInit, HandleLordEquipInit)
	//请求随机名称
	RegServerHandle(proto.Lord, proto.LordSRandomName, HandleLordRandomName)

	//更换装饰
	RegCommonMsg(sendDecorChange)
	//装备强化
	RegCommonMsg(sendLordEquipStreng)
	//领主转职
	RegCommonMsg(sendLordChangeJob)
	//学习天赋
	RegCommonMsg(sendLordTalentLearn)
	//升级天赋
	RegCommonMsg(sendLordTalentUpgrade)
	//领取VIP奖励
	RegCommonMsg(sendLordGetVipAwards)
	//查看领主信息
	RegCommonMsg(sendLordLookupLord)
	//查看英雄
	RegCommonMsg(sendLordLookupHero)
	//请求随机名称
	RegCommonMsg(sendLordRandomName)
	//技能进阶
	RegCommonMsg(sendLordSkillStage)
	//技能升级
	RegCommonMsg(sendLordSkillUpgrade)
	//技能位置
	RegCommonMsg(sendLordSkillExchangePos)
	//反馈
	RegCommonMsg(sendFeedback)
}

func HandleLordDecorInit(account *Account, reader *bytes.Reader) {
	decors := GetLordDecors(account)
	var l int16
	pack.Read(reader, &l)
	for i := int16(0); i < l; i++ {
		var (
			t     int
			id    int
			count int16
		)
		pack.Read(reader, &t, &id, &count)
		decor := &LordDecorData{id: id, unLock: make(map[int]bool)}
		decors[t] = decor
		for j := int16(0); j < count; j++ {
			var value int
			pack.Read(reader, &value)
			decor.unLock[value] = true
		}
	}
}

func HandleLordDecorUnlock(account *Account, reader *bytes.Reader) {
	var t int
	var id int
	pack.Read(reader, &t, &id)
	decors := GetLordDecors(account)
	decor := decors[t]
	decor.unLock[id] = true
}

func sendDecorChange(account *Account) {
	decors := GetLordDecors(account)
	for t, decor := range decors {
		if len(decor.unLock) == 0 {
			return
		}
		writer := pack.NewWriter()
		for id := range decor.unLock {
			pack.Write(writer, t, id)
			break
		}

		account.send(proto.Lord, proto.LordCDecorChange, writer.Bytes())

		break
	}
}

func HandleLordEquipInit(account *Account, reader *bytes.Reader) {
	var strengPos int
	var l int16
	pack.Read(reader, &strengPos, &l)
	equips := GetLordEquips(account)
	for i := int16(1); i <= l; i++ {
		equip := &LordEquipData{}
		pack.Read(reader, &equip.stage, &equip.level)
		equips[int(i)] = equip
	}
}

func HandleLordRandomName(account *Account, reader *bytes.Reader) {
	var (
		code int
		name string
	)
	pack.Read(reader, &code)
	if code != 0 {
		return
	}
	pack.Read(reader, &name)

	account.send(proto.Lord, proto.LordCChangeName, name)
}

func sendLordEquipStreng(account *Account) {
	account.send(proto.Lord, proto.LordCEquipStreng)
}

func sendLordChangeJob(account *Account) {
	account.send(proto.Lord, proto.LordCChangeJob, base.Rand(0, 10))
}

func sendLordTalentLearn(account *Account) {
	account.send(proto.Lord, proto.LordCTalentLearn, base.Rand(0, 35))
}

func sendLordTalentUpgrade(account *Account) {
	account.send(proto.Lord, proto.LordCTalentUpgrade, base.Rand(0, 35))
}

func sendLordGetVipAwards(account *Account) {
	account.send(proto.Lord, proto.LordCGetVipAwards, base.Rand(0, 12))
}

func sendLordLookupLord(account *Account) {
	for actorId, serverId := range serverActors {
		account.send(proto.Lord, proto.LordCLookupLord, 0, "", serverId, float64(actorId), "")
		break
	}
}

func sendLordLookupHero(account *Account) {
	for actorId, serverId := range serverActors {
		account.send(proto.Lord, proto.LordCLookupHero, 0, "", serverId, float64(actorId), base.Rand(0, 20))
		break
	}
}

func sendLordRandomName(account *Account) {
	account.send(proto.Lord, proto.LordCRandomName)
}

func sendLordSkillStage(account *Account) {
	account.send(proto.Lord, proto.LordCSkillStage, base.Rand(0, 5), base.Rand(0, 5))
}

func sendLordSkillUpgrade(account *Account) {
	count := int16(base.Rand(1, 5))
	writer := pack.NewWriter(count)
	for i := int16(1); i <= count; i++ {
		pack.Write(writer, int(i), base.Rand(1, 20))
	}
	account.send(proto.Lord, proto.LordCSkillUpgrade, writer.Bytes())
}

func sendLordSkillExchangePos(account *Account) {
	account.send(proto.Lord, proto.LordCSkillExchangePos, base.Rand(1, 5), base.Rand(1, 5))
}

func sendFeedback(account *Account) {
	account.send(proto.Base, proto.BaseCFeedback, int(0), "test feedback")
}
