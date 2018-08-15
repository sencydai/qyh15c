package main

import (
	"github.com/sencydai/gameworld/base"
	proto "github.com/sencydai/gameworld/proto/protocol"
)

func init() {
	//请求排行榜
	RegCommonMsg(sendRankData)
}

func sendRankData(account *Account) {
	index := base.Rand(1, 7)
	var rankName string
	switch index {
	case 1:
		rankName = "rank_ladder"
	case 2:
		rankName = "rankdb_fight"
	case 3:
		rankName = "rankdb_level"
	case 4:
		rankName = "rank_guild"
	case 5:
		rankName = "honorRoad"
	case 6:
		rankName = "legendBattle"
	case 7:
		rankName = "courageChallenge"
	}

	account.send(proto.Rank, proto.RankCRankData, rankName)
}
