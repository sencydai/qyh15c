package main

import (
	"reflect"
	"runtime/debug"
	"sync"
	"time"

	"github.com/sencydai/gameworld/base"
)

var (
	sysTimers     = make(map[string]*time.Timer)
	accountTimers = make(map[*Account]map[string]*time.Timer)
	timerLock     sync.RWMutex
)

func addTimer(account *Account, name string, delay int) *time.Timer {
	timerLock.Lock()
	defer timerLock.Unlock()

	sec := time.Second * time.Duration(delay)

	if account == nil {
		if t, ok := sysTimers[name]; ok {
			t.Stop()
		}
		t := time.NewTimer(sec)
		sysTimers[name] = t
		return t
	}

	accounts, ok := accountTimers[account]
	if !ok {
		accounts = make(map[string]*time.Timer)
		accountTimers[account] = accounts
	} else if t, ok := accounts[name]; ok {
		t.Stop()
	}

	t := time.NewTimer(sec)
	accounts[name] = t
	return t
}

func IsStoped(account *Account, name string) bool {
	timerLock.RLock()
	defer timerLock.RUnlock()

	if account == nil {
		_, ok := sysTimers[name]
		return !ok
	}

	accounts, ok := accountTimers[account]
	if !ok {
		return true
	}
	_, ok = accounts[name]
	return !ok
}

func StopTimer(account *Account, name string) bool {
	timerLock.Lock()
	defer timerLock.Unlock()

	if account == nil {
		t, ok := sysTimers[name]
		if !ok {
			return false
		}
		t.Stop()
		delete(sysTimers, name)
		return true
	}

	accounts, ok := accountTimers[account]
	if !ok {
		return false
	}
	t, ok := accounts[name]
	if !ok {
		return false
	}
	t.Stop()
	delete(accounts, name)
	return true
}

func StopAccountTimers(account *Account) {
	timerLock.Lock()
	defer timerLock.Unlock()

	accounts, ok := accountTimers[account]
	if !ok {
		return
	}
	for _, t := range accounts {
		t.Stop()
	}
	delete(accountTimers, account)
}

func callback(account *Account, cbFunc interface{}, args []interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf(account, "%v: %s", err, string(debug.Stack()))
		}
	}()

	cb, values := base.ReflectFunc(cbFunc, args)
	if account == nil {
		cb.Call(values)
	} else {
		v := []reflect.Value{reflect.ValueOf(account)}
		if len(values) > 0 {
			v = append(v, values...)
		}
		cb.Call(v)
	}
}

func After(account *Account, name string, delay int, cbFunc interface{}, args ...interface{}) {
	go func() {
		t := addTimer(account, name, delay)
		select {
		case <-t.C:
			StopTimer(account, name)
			callback(account, cbFunc, args)
		}
	}()
}

func Loop(account *Account, name string, delay, interval, times int, cbFunc interface{}, args ...interface{}) {
	go func() {
		loop := time.Second * time.Duration(interval)
		t := addTimer(account, name, delay)

		var count int
		for !IsStoped(account, name) {
			select {
			case <-t.C:
				if times > 0 {
					if count < times {
						count++
						if count == times {
							StopTimer(account, name)
							callback(account, cbFunc, args)
							return
						}

						t.Reset(loop)
						callback(account, cbFunc, args)
					}
				} else {
					t.Reset(loop)
					callback(account, cbFunc, args)
				}
			}
		}
	}()
}
