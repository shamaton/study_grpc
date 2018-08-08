package main

import (
	"errors"
	"flag"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/shamaton/study_grpc_go/chat/proto"
)

var debug = flag.Bool("debug", false, "debug mode")

const sessionIDLength = 16

type sessionID [sessionIDLength]byte

func alignSessionID(raw []byte) sessionID {
	var sid sessionID
	for i := 0; i < sessionIDLength && i < len(raw); i++ {
		sid[i] = raw[i]
	}
	return sid
}

type chatServer struct {
	mu   sync.RWMutex
	name map[sessionID]string
	buf  map[sessionID]chan *pb.Event

	last time.Time
	in   int64
	out  int64
}

func newChatServer() *chatServer {
	return &chatServer{
		name: make(map[sessionID]string),
		buf:  make(map[sessionID]chan *pb.Event),

		last: time.Now(),
	}
}

func (cs *chatServer) withReadLock(f func()) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	f()
}

func (cs *chatServer) withWriteLock(f func()) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	f()
}

func (cs *chatServer) generateSessionID() sessionID {
	var sid sessionID
	for i := 0; i < sessionIDLength/4; i++ {
		r := rand.Uint32()
		for j := 0; j < 4 && i*4+j < sessionIDLength; j++ {
			sid[i*4+j] = byte(r)
			r >>= 8
		}
	}
	return sid
}

func (cs *chatServer) unsafeExpire(sid sessionID) {
	if buf, ok := cs.buf[sid]; ok {
		close(buf)
	}
	delete(cs.name, sid)
	delete(cs.buf, sid)
}

func (cs *chatServer) Authorize(ctx context.Context, req *pb.RequestAuthorize) (*pb.ResponseAuthorize, error) {
	cs.in++

	if len(req.Name) == 0 {
		return nil, errors.New("name must be not empty")
	}
	if len(req.Name) > 20 {
		return nil, errors.New("name must be less than or equal 20 characters")
	}

	sid := cs.generateSessionID()
	cs.withWriteLock(func() {
		cs.name[sid] = req.Name
	})
	go func() {
		time.Sleep(5 * time.Second)
		cs.withWriteLock(func() {
			if _, ok := cs.buf[sid]; ok {
				return
			}
			cs.unsafeExpire(sid)
		})
	}()

	res := pb.ResponseAuthorize{
		SessionId: sid[:],
	}
	return &res, nil
}

func (cs *chatServer) Connect(req *pb.RequestConnect, stream pb.Chat_ConnectServer) error {
	cs.in++

	var (
		sid  sessionID      = alignSessionID(req.SessionId)
		buf  chan *pb.Event = make(chan *pb.Event, 1000)
		err  error
		name string
	)

	cs.withWriteLock(func() {
		var ok bool
		name, ok = cs.name[sid]
		if !ok {
			err = errors.New("not authorized")
			return
		}
		if _, ok := cs.buf[sid]; ok {
			err = errors.New("already connected")
			return
		}
		cs.buf[sid] = buf
	})
	if err != nil {
		return err
	}

	go cs.withReadLock(func() {
		log.Printf("Join name=%s", name)
		for _, buf := range cs.buf {
			buf <- &pb.Event{
				Event: &pb.Event_Join{
					Join: &pb.EventJoin{Name: name},
				},
			}
		}
	})

	// defer
	// 1: expire my session id
	// 2: send leave event to others
	defer cs.withReadLock(func() {
		log.Printf("Leave name=%s", name)
		for _, buf := range cs.buf {
			buf <- &pb.Event{
				Event: &pb.Event_Leave{
					Leave: &pb.EventLeave{Name: name},
				},
			}
		}
	})
	defer cs.withWriteLock(func() { cs.unsafeExpire(sid) })

	tick := time.Tick(time.Second)
	for {
		select {
		case <-stream.Context().Done():
			log.Println(name, "'s context done ...")
			return stream.Context().Err()
		case event := <-buf:
			// get exit event from myself
			if event.GetExit() != nil {
				log.Println(name, "'s stream close...")
				return nil
			}
			if err := stream.Send(event); err != nil {
				return err
			}
			cs.out++
		case <-tick:
			if err := stream.Send(&pb.Event{Event: &pb.Event_None{None: &pb.EventNone{}}}); err != nil {
				return err
			}
			cs.out++
		}
	}
}

func (cs *chatServer) Say(ctx context.Context, req *pb.CommandSay) (*pb.None, error) {
	cs.in++

	var (
		sid  sessionID = alignSessionID(req.SessionId)
		name string
		err  error
	)

	cs.withReadLock(func() {
		var ok bool
		name, ok = cs.name[sid]
		if !ok {
			err = errors.New("not authorized")
			return
		}
		if _, ok := cs.buf[sid]; !ok {
			err = errors.New("not authorized")
			return
		}
	})
	if err != nil {
		return nil, err
	}

	if len(req.Message) == 0 {
		return nil, errors.New("message must be not empty")
	}
	if len(req.Message) > 140 {
		return nil, errors.New("message must be less than or equal 140 characters")
	}

	go cs.withReadLock(func() {
		log.Printf("Log name=%s message=%s", name, req.Message)
		for _, buf := range cs.buf {
			buf <- &pb.Event{
				Event: &pb.Event_Log{
					Log: &pb.EventLog{
						Name:    name,
						Message: req.Message,
					},
				},
			}
		}
	})

	return &pb.None{}, nil
}

func (cs *chatServer) Leave(ctx context.Context, req *pb.CommandLeave) (*pb.None, error) {
	cs.in++

	var (
		sid  sessionID = alignSessionID(req.SessionId)
		name string
		err  error
	)

	defer cs.withWriteLock(func() {
		if _, ok := cs.name[sid]; !ok {
			err = errors.New("not authorized")
			return
		}
		if _, ok := cs.buf[sid]; !ok {
			err = errors.New("not found sid")
			return
		}

		// send exit event to myself
		cs.buf[sid] <- &pb.Event{
			Event: &pb.Event_Exit{Exit: &pb.EventExit{}},
		}
	})

	cs.withReadLock(func() {
		var ok bool
		name, ok = cs.name[sid]
		if !ok {
			err = errors.New("not authorized")
			return
		}
		if _, ok := cs.buf[sid]; !ok {
			err = errors.New("not found sid")
			return
		}
	})

	if err != nil {
		return nil, err
	}

	return &pb.None{}, err
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatalln("net.Listen:", err)
	}
	cs := newChatServer()
	go func() {
		tick := time.Tick(time.Second * 3)
		for {
			select {
			case <-tick:
				now := time.Now()
				duration := time.Now().Sub(cs.last)
				log.Printf("IN=%d OUT=%d IPS=%.0f OPS=%.0f auth=%d connect=%d",
					cs.in,
					cs.out,
					float64(cs.in)/float64(duration)*float64(time.Second),
					float64(cs.out)/float64(duration)*float64(time.Second),
					len(cs.name),
					len(cs.buf),
				)
				cs.last = now
				cs.in = 0
				cs.out = 0
			}
		}
	}()
	server := grpc.NewServer()
	pb.RegisterChatServer(server, cs)
	if err := server.Serve(lis); err != nil {
		log.Fatalln("Serve:", err)
	}
}
