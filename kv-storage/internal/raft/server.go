package raft

import (
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "kvstorage/api/proto"
	"kvstorage/internal/storage"
)

type PeerAddr string

type RaftServer struct {
	// ----- RAFT -----
	// Node state
	id    int64
	state Role
	peers []PeerAddr

	// persistentState
	curLeader int64
	curTerm   int64
	votedFor  *int64
	log       []LogEntry

	// volatileState
	commitIndex int64
	lastApplied int64

	// leaderState
	nextIndex  map[PeerAddr]int64
	matchIndex map[int64]int

	// Timers
	electionTimer   *time.Timer
	electionTimeout time.Duration

	heartbeatTimer   *time.Timer
	heartbeatTimeout time.Duration

	// ----- SERVICE -----
	pb.UnimplementedRaftServer

	mu      sync.Mutex
	storage *storage.LocalStorage
	logger  *zap.Logger
}

func NewRaftServer(myID int64, peers []PeerAddr, storage *storage.LocalStorage, logger *zap.Logger) *RaftServer {
	server := RaftServer{
		id:          myID,
		state:       Follower,
		peers:       peers,
		curTerm:     0,
		votedFor:    nil,
		log:         make([]LogEntry, 0),
		commitIndex: 0,
		lastApplied: 0,
		nextIndex:   make(map[PeerAddr]int64),
		matchIndex:  make(map[int64]int),

		electionTimeout:  time.Duration(8*myID+3) * time.Second,
		heartbeatTimeout: 2 * time.Second,

		logger:  logger,
		storage: storage,
	}

	server.resetElectionTimer()

	return &server
}

func (s *RaftServer) Start(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		s.logger.Fatal("failed to listen", zap.Error(err), zap.Int64("node_id", s.id))
		return
	}

	server := grpc.NewServer()
	pb.RegisterRaftServer(server, s)

	if err := server.Serve(listener); err != nil {
		s.logger.Fatal("failed to serve", zap.Error(err), zap.Int64("node_id", s.id))
	}

	s.logger.Info("GRPC server started", zap.String("port", port), zap.Int64("node_id", s.id))
}

func (s *RaftServer) IsLeader() bool {
	return s.curLeader == s.id
}

func (s *RaftServer) ReplicateEntry(entry LogEntry) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	var (
		prevLogInd  = int64(len(s.log) - 1)
		prevLogTerm int64
	)
	if prevLogInd > -1 {
		prevLogTerm = s.log[prevLogInd].Term
	}

	entry.Term = prevLogTerm
	s.log = append(s.log, entry)

	count := 1
	wg := sync.WaitGroup{}
	for _, peer := range s.peers {
		wg.Add(1)
		go func(peer PeerAddr) {
			defer wg.Done()

			req := &pb.AppendEntriesRequest{
				Term:         s.curTerm,
				LeaderID:     s.id,
				PrevLogIndex: prevLogInd,
				PrevLogTerm:  prevLogTerm,
				Entries:      []*pb.LogEntry{toPbEntry(&entry)},
				LeaderCommit: s.commitIndex,
			}

			res, err := sendAppendEntries(peer, req)
			if err != nil {
				s.logger.Error("Append entry error", zap.String("peer_id", string(peer)), zap.Error(err), zap.Int64("node_id", s.id))
				return
			}
			if !res.Success {
				s.logger.Debug("Append entry failed", zap.String("peer_id", string(peer)), zap.Int64("node_id", s.id))
				return
			}

			s.mu.Lock()

			s.nextIndex[peer]++
			count++

			s.mu.Unlock()
		}(peer)
	}
	wg.Wait()

	if count <= len(s.peers)/2 {
		return false
	}

	s.apply(entry)
	s.commitIndex = int64(len(s.log) - 1)

	return true
}
