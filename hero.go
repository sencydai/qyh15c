package main

import (
	"bytes"

	"github.com/sencydai/gameworld/base"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
)

type HeroData struct {
	army *HeroArmyData
}

type HeroArmyData struct {
	fight  map[int]int
	assist map[int]int
}

func GetHeroData(account *Account) *HeroData {
	if data, ok := account.data[keyHero]; ok {
		return data.(*HeroData)
	}
	account.data[keyHero] = &HeroData{}
	return account.data[keyHero].(*HeroData)
}

func GetHeroArmy(account *Account) *HeroArmyData {
	heroData := GetHeroData(account)
	if heroData.army == nil {
		heroData.army = &HeroArmyData{fight: make(map[int]int), assist: make(map[int]int)}
	}
	return heroData.army
}

func init() {
	//部队初始化
	RegServerHandle(proto.Hero, proto.HeroSArmyInit, HandleArmyInit)

	//设置部队英雄位置
	RegCommonMsg(sendSetArmyHeroPos)
	//一键升级
	RegCommonMsg(sendHeroOneKeyUpgrade)
	//英雄升阶
	RegCommonMsg(sendHeroUpgradeStage)
	//穿着装备
	RegCommonMsg(sendHeroWearEquip)
	//装备分解
	RegCommonMsg(sendHeroResolveEquip)
	//装备重铸
	RegCommonMsg(sendHeroRecastEquip)
	//穿着神器
	RegCommonMsg(sendHeroWearArti)
	//神器强化
	RegCommonMsg(sendHeroStrengArti)
	//英雄遣散
	RegCommonMsg(sendHeroDismiss)
	//英雄重修
	RegCommonMsg(sendHeroRebuild)
	//神器分解
	RegCommonMsg(sendHeroResolveArti)
}

func HandleArmyInit(account *Account, reader *bytes.Reader) {
	army := GetHeroArmy(account)
	army.fight = make(map[int]int)
	army.assist = make(map[int]int)
	var l int16
	pack.Read(reader, &l)
	for i := int16(0); i < l; i++ {
		var pos int
		var guid int
		pack.Read(reader, &pos, &guid)
		army.fight[pos] = guid
	}

	pack.Read(reader, &l)
	for i := int16(0); i < l; i++ {
		var pos int
		var guid int
		pack.Read(reader, &pos, &guid)
		army.assist[pos] = guid
	}
}

func sendSetArmyHeroPos(account *Account) {
	for guid := range GetBagHeros(account) {
		account.send(proto.Hero, proto.HeroCSetArmyHeroPos,
			guid, base.Rand(1, 2), base.Rand(1, 12))
		break
	}
}

func sendHeroOneKeyUpgrade(account *Account) {
	for guid := range GetBagHeros(account) {
		account.send(proto.Hero, proto.HeroCOneKeyUpgrade, guid)
		break
	}
}

func sendHeroUpgradeStage(account *Account) {
	for guid := range GetBagHeros(account) {
		account.send(proto.Hero, proto.HeroCUpgradeStage, guid)
		break
	}
}

func sendHeroChangeJob(account *Account) {
	for guid, hero := range GetBagHeros(account) {
		account.send(proto.Hero, proto.HeroCUpgradeStage,
			guid, hero.id+base.Rand(-100, 200))
		break
	}
}

func sendHeroWearEquip(account *Account) {
	for guid := range GetBagEquips(account) {
		account.send(proto.Hero, proto.HeroCWearEquip,
			base.Rand(1, 6), guid, base.Rand(0, 1))
		break
	}
}

func sendHeroStrengEquip(account *Account) {
	for guid := range GetBagEquips(account) {
		account.send(proto.Hero, proto.HeroCStrengEquip,
			guid, base.Rand(0, 1))
		break
	}
}

func sendHeroResolveEquip(account *Account) {
	equips := GetBagEquips(account)
	count := base.Rand(0, len(equips))
	writer := pack.NewWriter(int16(count))
	var i int
	for guid := range equips {
		if i < count {
			pack.Write(writer, guid)
			i++
		} else {
			break
		}
	}
	account.send(proto.Hero, proto.HeroCResolveEquip, writer.Bytes())
}

func sendHeroRecastEquip(account *Account) {
	equips := GetBagEquips(account)
	count := base.Rand(0, len(equips))
	writer := pack.NewWriter(int16(count))
	var i int
	for guid := range equips {
		if i < count {
			pack.Write(writer, guid)
			i++
		} else {
			break
		}
	}
	account.send(proto.Hero, proto.HeroCRecastEquip, writer.Bytes())
}

func sendHeroWearArti(account *Account) {
	for guid := range GetBagArtis(account) {
		account.send(proto.Hero, proto.HeroCWearArti,
			base.Rand(1, 6), guid, base.Rand(0, 1))
		break
	}
}

func sendHeroStrengArti(account *Account) {
	for guid := range GetBagArtis(account) {
		account.send(proto.Hero, proto.HeroCStrengArti, guid)
		break
	}
}

func sendHeroDismiss(account *Account) {
	heros := GetBagHeros(account)
	count := base.Rand(0, len(heros))
	writer := pack.NewWriter(int16(count))
	var i int
	for guid := range heros {
		if i < count {
			pack.Write(writer, guid)
			i++
		} else {
			break
		}
	}
	account.send(proto.Hero, proto.HeroCHeroDismiss, writer.Bytes())
}

func sendHeroRebuild(account *Account) {
	heros := GetBagHeros(account)
	count := base.Rand(0, len(heros))
	writer := pack.NewWriter(int16(count))
	var i int
	for guid := range heros {
		if i < count {
			pack.Write(writer, guid)
			i++
		} else {
			break
		}
	}
	account.send(proto.Hero, proto.HeroCHeroRebuild, writer.Bytes())
}

func sendHeroResolveArti(account *Account) {
	artis := GetBagArtis(account)
	count := base.Rand(0, len(artis))
	writer := pack.NewWriter(int16(count))
	var i int
	for guid := range artis {
		if i < count {
			pack.Write(writer, guid)
			i++
		} else {
			break
		}
	}
	account.send(proto.Hero, proto.HeroCResolveArti, writer.Bytes())
}
