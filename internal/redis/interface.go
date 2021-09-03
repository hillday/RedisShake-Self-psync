package redis

import (
	"bufio"
	"github.com/alibaba/RedisShake/internal/log"
)

type Redis interface {
	DoWithStringReply(args ...string) string
	Send(args ...string)
	Flush()

	Receive() (interface{}, error)

	BufioReader() *bufio.Reader
	SetBufioReader(reader *bufio.Reader)
}

func ArrayString(replyInterface interface{}, err error) []string {
	if err != nil {
		log.PanicError(err)
	}
	replyArray := replyInterface.([]interface{})
	replyArrayString := make([]string, len(replyArray))
	for inx, item := range replyArray {
		replyArrayString[inx] = item.(string)
	}
	return replyArrayString
}
