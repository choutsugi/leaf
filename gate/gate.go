package gate

import (
	"github.com/choutsugi/leaf/chanrpc"
	"github.com/choutsugi/leaf/log"
	"github.com/choutsugi/leaf/network"
	"github.com/xtaci/kcp-go"
	"net"
	"reflect"
	"time"
)

type Gate struct {
	MaxConnNum      int
	PendingWriteNum int

	Processor    network.Processor
	AgentChanRPC *chanrpc.Server

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string

	//udp
	UDPAddr string

	//tcp
	TCPAddr string

	// msg parser
	LenMsgLen    int
	MinMsgLen    uint32
	MaxMsgLen    uint32
	LittleEndian bool
	MsgParser    *network.UDPMsgParser // TODO：how to be compatible with tcp
}

func (gate *Gate) Run(closeSig chan bool) {

	// TODO：WSServer

	// TODO：TCPServer

	var udpServer *network.UDPServer
	if gate.UDPAddr != "" {
		udpServer = new(network.UDPServer)
		udpServer.Addr = gate.UDPAddr
		udpServer.MaxConnNum = gate.MaxConnNum

		udpServer.NewAgent = func(conn *kcp.UDPSession) network.Agent {
			a := &agent{conn: conn, gate: gate}
			if gate.AgentChanRPC != nil {
				gate.AgentChanRPC.Go("NewAgent", a)
			}
			return a
		}
		// msg parser
		msgParser := network.NewUDPMsgParser()
		msgParser.SetMsgLen(gate.LenMsgLen, gate.MinMsgLen, gate.MaxMsgLen)
		msgParser.SetByteOrder(gate.LittleEndian)
		gate.MsgParser = msgParser

	}

	if udpServer != nil {
		udpServer.Start()
	}

	<-closeSig

	if udpServer != nil {
		udpServer.Close()
	}
}

func (gate *Gate) OnDestroy() {}

type agent struct {
	conn     *kcp.UDPSession
	gate     *Gate
	userData interface{}
}

func (a *agent) Run() {
	buf := make([]byte, 1024*10)
	for {
		a.conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		// read data
		n, err := a.conn.Read(buf)
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}
		log.Debug("a.conn.Read(buf): %v", n)

		// parse data
		msgData, err := a.gate.MsgParser.ReadParse(buf[:n])
		if err != nil {
			log.Debug("read message UDPMsgParser ReadParse: %v", err)
			break
		}

		if a.gate.Processor != nil {
			msg, err := a.gate.Processor.Unmarshal(msgData)
			if err != nil {
				log.Debug("unmarshal message error: %v", err)
				break
			}
			err = a.gate.Processor.Route(msg, a)
			if err != nil {
				log.Debug("route message error: %v", err)
				break
			}
		}
	}
}

func (a *agent) OnClose() {
	if a.gate.AgentChanRPC != nil {
		err := a.gate.AgentChanRPC.Open(0).Call0("CloseAgent", a)
		if err != nil {
			log.Error("chanrpc error: %v", err)
		}
	}
}

func (a *agent) WriteMsg(msg interface{}) {
	switch msg.(type) {
	case string:
		msgData, err := a.gate.MsgParser.WriteParse([]byte(msg.(string)))
		if err != nil {
			log.Error("string message UDPMsgParser WriteParse: %v", err)
			return
		}
		a.conn.Write(msgData)
	default:
		if a.gate.Processor != nil {
			data, err := a.gate.Processor.Marshal(msg)
			if err != nil {
				log.Error("marshal message %v error: %v", reflect.TypeOf(msg), err)
				return
			}
			msgData, err := a.gate.MsgParser.WriteParse(data)
			if err != nil {
				log.Error("marshal message UDPMsgParser WriteParse: %v", err)
				return
			}
			a.conn.Write(msgData)
		}
	}
}

func (a *agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) Destroy() {
	// TODO：how to process kcp destroy
	//a.conn.Destroy()
}

func (a *agent) UserData() interface{} {
	return a.userData
}

func (a *agent) SetUserData(data interface{}) {
	a.userData = data
}
