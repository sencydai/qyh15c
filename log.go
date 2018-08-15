package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sencydai/gameworld/base"
)

type logger struct {
	file   *os.File
	writer *bufio.Writer
	lock   sync.Mutex
}

var (
	log = &logger{file: os.Stdout, writer: bufio.NewWriterSize(os.Stdout, 1024*10)}
)

func init() {
	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 100):
				log.sync()
			}
		}
	}()
}

func (l *logger) sync() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.writer.Flush()
	l.file.Sync()
}

func (l *logger) Close() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.writer.Flush()
	l.file.Sync()
}

func (l *logger) Print(account *Account, data ...interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	var text string
	if account != nil {
		text = fmt.Sprintf("%s [%s] [%s,%d,%d] - %s\n",
			base.FormatDateTime(time.Now()), base.FileLine(2),
			account.accountName, account.accountId, account.actorId,
			fmt.Sprint(data...),
		)
	} else {
		text = fmt.Sprintf("%s [%s] - %s\n",
			base.FormatDateTime(time.Now()), base.FileLine(2),
			fmt.Sprint(data...),
		)
	}

	l.writer.WriteString(text)
}

func (l *logger) Printf(account *Account, format string, data ...interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	var text string
	if account != nil {
		text = fmt.Sprintf("%s [%s] [%s,%d,%d] - %s\n",
			base.FormatDateTime(time.Now()), base.FileLine(2),
			account.accountName, account.accountId, account.actorId,
			fmt.Sprintf(format, data...),
		)
	} else {
		text = fmt.Sprintf("%s [%s] - %s\n",
			base.FormatDateTime(time.Now()), base.FileLine(2),
			fmt.Sprintf(format, data...),
		)
	}

	l.writer.WriteString(text)
}
