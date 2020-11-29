package events

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/xo/dburl"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/omecodes/zebou"
	"github.com/siddontang/go-mysql/canal"
)

type Server struct {
	canal.DummyEventHandler
	sync.Mutex

	config *Config

	errors        chan error
	stopRequested bool
	dbCanal       *canal.Canal
	hub           *zebou.Hub
	peers         []*zebou.PeerInfo
	listener      net.Listener
}

func (s *Server) OnRow(e *canal.RowsEvent) error {
	fmt.Println(e.String())
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
	return s.config.Address
}

func (s *Server) start() {
	if !strings.HasPrefix(s.config.DBUri, "mysql://") {
		s.config.DBUri = "mysql://" + s.config.DBUri
	}

	u, err := dburl.Parse(s.config.DBUri)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		c := canal.NewDefaultConfig()
		c.Addr = fmt.Sprintf("%s:%s", u.Host, u.Port())
		c.User = u.User.Username()
		c.Password, _ = u.User.Password()
		c.Dump.TableDB = u.DSN
		c.Dump.Tables = []string{"objects"}

		s.dbCanal, err = canal.NewCanal(c)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		}

		s.dbCanal.SetEventHandler(s)

		err = s.dbCanal.Run()
		if err != nil {
			fmt.Println(err)
			<-time.After(time.Second)
		}
	}
}

func (s *Server) Stop() error {
	s.stopRequested = true
	return s.listener.Close()
}

func (s *Server) Errors() chan error {
	if s.errors != nil {
		s.errors = make(chan error)
	}
	return s.errors
}

func Serve(cfg *Config) (*Server, error) {
	s := &Server{
		config: cfg,
	}

	var err error
	if cfg.TlsConfig != nil {
		s.listener, err = tls.Listen("tcp", cfg.Address, cfg.TlsConfig)
	} else {
		s.listener, err = net.Listen("tcp", cfg.Address)
	}
	if err != nil {
		return nil, err
	}

	s.hub, err = zebou.Serve(s.listener, s)
	if err != nil {
		return nil, err
	}

	go s.start()
	return s, nil
}
