package network

import (
	"github.com/choutsugi/forest/log"
	"github.com/xtaci/kcp-go"

	"sync"
)

type UDPServer struct {
	Addr       string
	MaxConnNum int
	NewAgent   func(*kcp.UDPSession) Agent
	ln         *kcp.Listener

	wgLn sync.WaitGroup
}

func (server *UDPServer) Start() {
	server.init()
	go server.run()
}

func (server *UDPServer) init() {

	ln, err := kcp.Listen(server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Release("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	// config listener
	server.ln = ln.(*kcp.Listener)
	server.ln.SetReadBuffer(4 * 1024 * 1024)
	server.ln.SetWriteBuffer(4 * 1024 * 1024)
	server.ln.SetDSCP(46)
}

func (server *UDPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	for {
		conn, err := server.ln.Accept()
		if err != nil {
			log.Release("accept error: %v", err)
			return
		}

		// TODO：kcp-go未提供获取已有连接数量的方法
		//if server.ln.SessionCount() >= server.MaxConnNum {
		//	conn.Close()
		//	log.Debug("too many connections")
		//	continue
		//}

		// config sess
		sess := conn.(*kcp.UDPSession)
		sess.SetReadBuffer(4 * 1024 * 1024)
		sess.SetWriteBuffer(4 * 1024 * 1024)
		sess.SetWindowSize(4096, 4096)
		sess.SetWriteDelay(true)
		sess.SetACKNoDelay(false)
		sess.SetNoDelay(1, 100, 2, 0)

		agent := server.NewAgent(sess)

		go func() {
			agent.Run()
			agent.OnClose()

		}()
	}
}

func (server *UDPServer) Close() {
	server.ln.Close()
	server.wgLn.Wait()
}
