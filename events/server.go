package events

import (
	"context"
	"crypto/tls"
	"net"
	"sync"

	"github.com/omecodes/omestore/common"
	"github.com/omecodes/zebou/v2"
	"github.com/siddontang/go-mysql/canal"
)

type Server struct {
	canal.DummyEventHandler
	sync.Mutex
	dsn       string
	address   string
	validator common.CredentialsValidator
	tc        *tls.Config
	dbCanal   *canal.Canal
	hub       *zebou.Hub
	peers     []*zebou.PeerInfo
	listener  net.Listener
	err       error
}

func (s *Server) OnRow(e *canal.RowsEvent) error {
	return nil
}

func (s *Server) String() string {
	return ""
}

func (s *Server) NewClient(ctx context.Context, info *zebou.PeerInfo) {
	s.Lock()
	defer s.Unlock()

	s.peers = append(s.peers, info)
}

func (s *Server) ClientQuit(ctx context.Context, info *zebou.PeerInfo) {
	s.Lock()
	defer s.Unlock()

	var newPeersList []*zebou.PeerInfo
	for _, in := range s.peers {
		if in.ID != info.ID {
			newPeersList = append(newPeersList, in)
		}
	}
	s.peers = newPeersList
}

func (s *Server) OnMessage(ctx context.Context, msg *zebou.ZeMsg) {
	//peer := zebou.Peer(ctx)
}

func (s *Server) Address() string {
	return s.address
}

func (s *Server) start() error {

	s.hub, s.err = zebou.Serve(s.listener, s)
	if s.err != nil {
		return s.err
	}
	return s.err
}

func (s *Server) Stop() error {
	return nil
}

func Serve(dsn string, address string, validator common.CredentialsValidator) (*Server, error) {
	s := &Server{
		dsn:       dsn,
		address:   address,
		validator: validator,
	}
	var err error
	s.listener, err = net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return s, s.start()
}

func ServeOverTLS(dsn string, address string, tc *tls.Config, validator common.CredentialsValidator) (*Server, error) {
	s := &Server{
		dsn:       dsn,
		address:   address,
		tc:        tc,
		validator: validator,
	}
	var err error
	s.listener, err = tls.Listen("tcp", address, tc)
	if err != nil {
		return nil, err
	}
	return s, s.start()
}
