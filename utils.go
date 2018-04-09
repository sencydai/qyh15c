package main

import (
	"fmt"
	"math/rand"
	"time"
)

func TimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func Printf(account *AccountData, format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf("%s [%s,%d] - %s",
		TimeFormat(time.Now()), account.accountName, int64(account.actorId), fmt.Sprintf(format, a...)))
}

func Print(account *AccountData, a ...interface{}) {
	fmt.Println(fmt.Sprintf("%s [%s,%d] - %s",
		TimeFormat(time.Now()), account.accountName, int64(account.actorId), fmt.Sprint(a...)))
}

func RandInt31n(start, end int32) int32 {
	if start >= end {
		return end
	}
	return start + rand.Int31n(end-start+1)
}
