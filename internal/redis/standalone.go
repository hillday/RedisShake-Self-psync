package redis

import (
	"bufio"
	"crypto/tls"
	"github.com/alibaba/RedisShake/internal/log"
	"github.com/alibaba/RedisShake/internal/redis/proto"
	"net"
	"time"
)

type Standalone struct {
	reader      *bufio.Reader
	writer      *bufio.Writer
	protoReader *proto.Reader
	protoWriter *proto.Writer
}

func NewStandalone(address string, password string, isTls bool) Redis {
	r := new(Standalone)
	var conn net.Conn
	var dialer net.Dialer
	var err error
	dialer.Timeout = 3 * time.Second
	if isTls {
		conn, err = tls.DialWithDialer(&dialer, "tcp", address, &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = dialer.Dial("tcp", address)
	}
	if err != nil {
		log.PanicError(err)
	}

	r.reader = bufio.NewReader(conn)
	r.writer = bufio.NewWriter(conn)
	r.protoReader = proto.NewReader(r.reader)
	r.protoWriter = proto.NewWriter(r.writer)

	// auth
	if password != "" {
		reply := r.DoWithStringReply("auth", password)
		if reply != "OK" {
			log.Panicf("auth failed with reply: %s", reply)
		}
		log.Infof("auth successful. address=[%s]", address)
	} else {
		log.Infof("no password. address=[%s]", address)
	}

	// ping to test connection
	reply := r.DoWithStringReply("ping")

	if reply != "PONG" {
		panic("ping failed with reply: " + reply)
	}

	return r
}

func (r *Standalone) SetBufioReader(rd *bufio.Reader) {
	r.reader = rd
	r.protoReader = proto.NewReader(r.reader)
}

func (r *Standalone) DoWithStringReply(args ...string) string {
	r.Send(args...)
	r.Flush()

	replyInterface, err := r.Receive()
	if err != nil {
		log.PanicError(err)
	}
	reply := replyInterface.(string)
	return reply
}

func (r *Standalone) Send(args ...string) {
	argsInterface := make([]interface{}, len(args))
	for inx, item := range args {
		argsInterface[inx] = item
	}
	err := r.protoWriter.WriteArgs(argsInterface)
	if err != nil {
		log.PanicError(err)
	}
}

func (r *Standalone) Flush() {
	err := r.writer.Flush()
	if err != nil {
		log.PanicError(err)
	}
}

func (r *Standalone) Receive() (interface{}, error) {
	return r.protoReader.ReadReply()
}

func (r *Standalone) BufioReader() *bufio.Reader {
	return r.reader
}
