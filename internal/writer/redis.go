package writer

import (
	"github.com/alibaba/RedisShake/internal/config"
	"github.com/alibaba/RedisShake/internal/entry"
	"github.com/alibaba/RedisShake/internal/log"
	redis "github.com/alibaba/RedisShake/internal/redis"
	"strconv"
)

type redisWriter struct {
	client    redis.Redis
	sendCount uint64 // have sent
	sendBytes uint64 // have sent in bytes
	DbId      int

	chWaitReply chan *entry.Entry
}

func NewRedisWriter(address string, password string, isTls bool) Writer {
	rw := new(redisWriter)
	rw.client = redis.NewStandalone(address, password, isTls)
	log.Infof("redisWriter connected to redis successful. address=[%s]", address)

	rw.chWaitReply = make(chan *entry.Entry, config.Config.Advanced.PipelineCountLimit)
	go rw.flushInterval()
	return rw
}

func (w *redisWriter) Write(e *entry.Entry) {
	// switch db if we need
	if w.DbId != e.DbId {
		w.switchDbTo(e.DbId)
	}

	// send
	log.Debugf("redisWriter send command. argv=%v", e.Argv)
	w.client.Send(e.Argv...)
	w.client.Flush()

	w.chWaitReply <- e
	//// statistics
	//w.sendCount += 1
	//w.sendBytes += uint64(len(entry.Argv))
	//for _, arg := range entry.Argv {
	//	w.sendBytes += uint64(len(arg))
	//}
	//
	//// Flush pipeline
	//if w.sendCount > config.Config.Advanced.PipelineCountLimit ||
	//	w.sendBytes > config.Config.Advanced.TargetRedisClientMaxQuerybufLen {
	//	w.chWaitReply <- entry
	//	w.sendBytes = 0
	//}
}

func (w *redisWriter) switchDbTo(newDbId int) {
	w.client.Send("select", strconv.Itoa(newDbId))
	w.DbId = newDbId
}

func (w *redisWriter) flushInterval() {
	for {
		select {
		case e := <-w.chWaitReply:
			reply, err := w.client.Receive()
			log.Debugf("redisWriter received reply. argv=%v, reply=%v, error=[%v]", e.Argv, reply, err)
			if err != nil {
				if err.Error() == "BUSYKEY Target key name already exists." {
					if config.Config.Advanced.RDBRestoreCommandBehavior == "skip" {
						log.Warnf("redisWriter received BUSYKEY reply. argv=%v", e.Argv)
					} else if config.Config.Advanced.RDBRestoreCommandBehavior == "panic" {
						log.Panicf("redisWriter received BUSYKEY reply. argv=%v", e.Argv)
					}
				} else {
					log.Panicf("redisWriter received error. error=[%v], argv=%v", err, e.Argv)
				}
			}
		}
	}
}
