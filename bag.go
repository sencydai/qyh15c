package main

import (
	"bytes"

	"github.com/sencydai/gameworld/base"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
)

type BagData struct {
	items  map[int]int
	heros  map[int]*BagHeroData
	equips map[int]*BagEquipData
	artis  map[int]*BagArtiData
}

type BagHeroData struct {
	guid    int
	posType byte
	pos     int16
	posMap  int
	id      int
	level   int16
	exp     int
	stage   int16
}

type BagEquipData struct {
	guid  int
	pos   int
	id    int
	level int
}

type BagArtiData struct {
	guid        int
	pos         int
	id          int
	attrs       []int
	strengLevel []int
	strengPos   int
}

func GetBagData(account *Account) *BagData {
	if data, ok := account.data[keyBag]; ok {
		return data.(*BagData)
	}

	account.data[keyBag] = &BagData{}
	return account.data[keyBag].(*BagData)
}

func GetBagItems(account *Account) map[int]int {
	bagData := GetBagData(account)
	if bagData.items == nil {
		bagData.items = make(map[int]int)
	}
	return bagData.items
}

func GetBagHeros(account *Account) map[int]*BagHeroData {
	bagData := GetBagData(account)
	if bagData.heros == nil {
		bagData.heros = make(map[int]*BagHeroData)
	}
	return bagData.heros
}

func GetBagEquips(account *Account) map[int]*BagEquipData {
	bagData := GetBagData(account)
	if bagData.equips == nil {
		bagData.equips = make(map[int]*BagEquipData)
	}
	return bagData.equips
}

func GetBagArtis(account *Account) map[int]*BagArtiData {
	bagData := GetBagData(account)
	if bagData.artis == nil {
		bagData.artis = make(map[int]*BagArtiData)
	}
	return bagData.artis
}

func init() {
	//物品初始化
	RegServerHandle(proto.Bag, proto.BagSItemInit, HandleBagItemInit)
	//物品删除
	RegServerHandle(proto.Bag, proto.BagSItemDelete, HandleBagItemDelete)
	//英雄初始化
	RegServerHandle(proto.Bag, proto.BagSHeroInit, HandleBagHeroInit)
	//英雄修改
	RegServerHandle(proto.Bag, proto.BagSHeroUpdate, HandleBagHeroUpdate)
	//英雄删除
	RegServerHandle(proto.Bag, proto.BagSHeroDelete, HandleBagHeroDelete)
	//装备初始化
	RegServerHandle(proto.Bag, proto.BagSEquipInit, HandleBagEquipInit)
	//装备修改
	RegServerHandle(proto.Bag, proto.BagSEquipUpdate, HandleBagEquipUpdate)
	//装备删除
	RegServerHandle(proto.Bag, proto.BagSEquipDelete, HandleBagEquipDelete)
	//神器初始化
	RegServerHandle(proto.Bag, proto.BagSArtiInit, HandleBagArtiInit)
	//神器修改
	RegServerHandle(proto.Bag, proto.BagSArtiUpdate, HandleBagArtiUpdate)
	//神器删除
	RegServerHandle(proto.Bag, proto.BagSArtiDelete, HandleBagArtiDelete)
	//货币初始化
	RegServerHandle(proto.Bag, proto.BagSCurrencyInit, HandleBagCurrencyInit)
	//货币删除
	RegServerHandle(proto.Bag, proto.BagSCurrencyDelete, HandleBagCurrencyDelete)
	//添加奖励
	RegServerHandle(proto.Bag, proto.BagSAddAwards, HandleBagAddAwards)

	//开启宝箱
	RegCommonMsg(sendOpenBox)
	//合成
	RegCommonMsg(sendCompose)
}

func HandleBagItemInit(account *Account, reader *bytes.Reader) {
	items := GetBagItems(account)
	var count int16
	pack.Read(reader, &count)
	for i := int16(0); i < count; i++ {
		var itemType int16
		var itemCount int16
		pack.Read(reader, &itemType, &itemCount)
		for j := int16(0); j < itemCount; j++ {
			var id int
			var total int
			pack.Read(reader, &id, &total)
			items[id] = total
		}
	}
}

func HandleBagCurrencyInit(account *Account, reader *bytes.Reader) {
	items := GetBagItems(account)
	var count int16
	pack.Read(reader, &count)
	for i := int16(0); i < count; i++ {
		var id int
		var total int
		pack.Read(reader, &id, &total)
		items[id] = total
	}
}

func deleteItem(account *Account, id, total, change int) {
	items := GetBagItems(account)
	if cur, ok := items[id]; ok {
		newCount := cur + change
		if newCount != total || newCount < 0 || total < 0 {
			log.Printf(account, "HandleBagItemDelete: id(%d),client(%d),server(%d),change(%d)", id, cur, total, change)
		}
	} else {
		log.Printf(account, "HandleBagItemDelete: not find item,id(%d),server(%d),change(%d)",
			id, total, change)
	}

	items[id] = total
	if items[id] <= 0 {
		delete(items, id)
	}
}

func HandleBagItemDelete(account *Account, reader *bytes.Reader) {
	var itemType int
	var id int
	var total int
	var change int

	pack.Read(reader, &itemType, &id, &total, &change)
	deleteItem(account, id, total, change)
}

func HandleBagCurrencyDelete(account *Account, reader *bytes.Reader) {
	var id int
	var total int
	var change int
	pack.Read(reader, &id, &total, &change)
	deleteItem(account, id, total, change)
}

func HandleBagHeroInit(account *Account, reader *bytes.Reader) {
	heros := GetBagHeros(account)
	var count int16
	pack.Read(reader, &count)
	for i := int16(0); i < count; i++ {
		hero := &BagHeroData{}
		pack.Read(reader, &hero.guid, &hero.posType, &hero.pos, &hero.posMap, &hero.id, &hero.level, &hero.exp, &hero.stage)
		heros[hero.guid] = hero
	}
}

func HandleBagHeroUpdate(account *Account, reader *bytes.Reader) {
	hero := &BagHeroData{}
	pack.Read(reader, &hero.guid, &hero.posType, &hero.pos, &hero.posMap, &hero.id, &hero.level, &hero.exp, &hero.stage)
	heros := GetBagHeros(account)
	if heros[hero.guid] == nil {
		log.Printf(account, "HandleBagHeroUpdate: not find hero(%d)", hero.guid)
	}
	heros[hero.guid] = hero
}

func HandleBagHeroDelete(account *Account, reader *bytes.Reader) {
	heros := GetBagHeros(account)
	var count int16
	pack.Read(reader, &count)
	for i := int16(0); i < count; i++ {
		var guid int
		pack.Read(reader, &guid)
		if heros[guid] == nil {
			log.Printf(account, "HandleBagHeroDelete: not find hero(%d)", guid)
		} else {
			delete(heros, guid)
		}
	}
}

func HandleBagEquipInit(account *Account, reader *bytes.Reader) {
	equips := GetBagEquips(account)
	var count int16
	pack.Read(reader, &count)
	for i := int16(0); i < count; i++ {
		equip := &BagEquipData{}
		pack.Read(reader, &equip.guid, &equip.pos, &equip.id, &equip.level)
		equips[equip.guid] = equip
	}
}

func HandleBagEquipUpdate(account *Account, reader *bytes.Reader) {
	equip := &BagEquipData{}
	pack.Read(reader, &equip.guid, &equip.pos, &equip.id, &equip.level)
	equips := GetBagEquips(account)
	if equips[equip.guid] == nil {
		log.Printf(account, "HandleBagEquipUpdate: not find equip(%d)", equip.guid)
	}
	equips[equip.guid] = equip
}

func HandleBagEquipDelete(account *Account, reader *bytes.Reader) {
	equips := GetBagEquips(account)
	var source uint8
	var count int16
	pack.Read(reader, &source, &count)
	for i := int16(0); i < count; i++ {
		var guid int
		pack.Read(reader, &guid)
		if equips[guid] == nil {
			log.Printf(account, "HandleBagEquipDelete: not find equip(%d)", guid)
		} else {
			delete(equips, guid)
		}
	}
}

func HandleBagArtiInit(account *Account, reader *bytes.Reader) {
	artis := GetBagArtis(account)
	var count int16
	pack.Read(reader, &count)
	for i := int16(0); i < count; i++ {
		arti := &BagArtiData{}
		var l int16
		pack.Read(reader, &arti.guid, &arti.pos, &arti.id, &l)
		arti.attrs = make([]int, l)
		for j := int16(0); j < l; j++ {
			var value int
			pack.Read(reader, &value)
			arti.attrs[j] = value
		}
		pack.Read(reader, &l)
		arti.strengLevel = make([]int, l)
		for j := int16(0); j < l; j++ {
			var value int
			pack.Read(reader, &value)
			arti.strengLevel[j] = value
		}
		pack.Read(reader, &arti.strengPos)
		artis[arti.guid] = arti
	}
}

func HandleBagArtiUpdate(account *Account, reader *bytes.Reader) {
	arti := &BagArtiData{}
	var l int16
	pack.Read(reader, &arti.guid, &arti.pos, &arti.id, &l)
	arti.attrs = make([]int, l)
	for j := int16(0); j < l; j++ {
		var value int
		pack.Read(reader, &value)
		arti.attrs[j] = value
	}
	pack.Read(reader, &l)
	arti.strengLevel = make([]int, l)
	for j := int16(0); j < l; j++ {
		var value int
		pack.Read(reader, &value)
		arti.strengLevel[j] = value
	}
	pack.Read(reader, &arti.strengPos)

	artis := GetBagArtis(account)
	if artis[arti.guid] == nil {
		log.Printf(account, "HandleBagArtiUpdate: not find arti(%d)", arti.guid)
	}
	artis[arti.guid] = arti
}

func HandleBagArtiDelete(account *Account, reader *bytes.Reader) {
	artis := GetBagArtis(account)
	var count int16
	pack.Read(reader, &count)
	for i := int16(0); i < count; i++ {
		var guid int
		pack.Read(reader, &guid)
		if artis[guid] == nil {
			log.Printf(account, "HandleBagArtiDelete: not find arti(%d)", guid)
		} else {
			delete(artis, guid)
		}
	}
}

func HandleBagAddAwards(account *Account, reader *bytes.Reader) {
	var source int
	var l int16
	pack.Read(reader, &source, &l)
	for i := int16(0); i < l; i++ {
		var awardType int
		var id int
		var addCount int
		pack.Read(reader, &awardType, &id, &addCount)
		switch awardType {
		//物品
		case 1:
			var totalCount int
			pack.Read(reader, &totalCount)
			items := GetBagItems(account)
			//领主经验 公会资金
			if id != 4 && id != 8 {
				if oldCount, ok := items[id]; ok && (oldCount+addCount) != totalCount {
					log.Printf(account, "HandleBagAddAwards: source(%d) item(%d) old(%d),new(%d),add(%d)",
						source, id, oldCount, totalCount, addCount)
				}

				items[id] = totalCount
			}

		//英雄
		case 7:
			heros := GetBagHeros(account)
			for j := 0; j < addCount; j++ {
				var guid int
				pack.Read(reader, &guid)
				heros[guid] = &BagHeroData{guid: guid, id: id, level: 1}
			}
		//英雄装备
		case 8:
			equips := GetBagEquips(account)
			for j := 0; j < addCount; j++ {
				var guid int
				pack.Read(reader, &guid)
				equips[guid] = &BagEquipData{guid: guid, id: id}
			}
		//神器
		case 9:
			artis := GetBagArtis(account)
			for j := 0; j < addCount; j++ {
				var guid int
				var attrLen int16
				pack.Read(reader, &guid, &attrLen)
				arti := &BagArtiData{guid: guid, id: id, strengLevel: []int{0, 0, 0, 0}, strengPos: 1}
				arti.attrs = make([]int, attrLen)
				for k := int16(0); k < attrLen; k++ {
					var value int
					pack.Read(reader, &value)
					arti.attrs[k] = value
				}
				artis[guid] = arti
			}
		}
	}
}

func sendOpenBox(account *Account) {
	items := GetBagItems(account)
	if len(items) == 0 {
		return
	}
	for key := range items {
		account.send(proto.Bag, proto.BagCOpenBox, key, base.Rand(0, 10))
		break
	}
}

func sendCompose(account *Account) {
	items := GetBagItems(account)
	if len(items) == 0 {
		return
	}
	count := base.Rand(0, len(items))
	writer := pack.NewWriter(int16(count))
	var i int
	for key := range items {
		pack.Write(writer, key, base.Rand(0, 10))
		i++
		if i == count {
			break
		}
	}

	account.send(proto.Bag, proto.BagCCompose, writer.Bytes())
}
