package raft

import "go.uber.org/zap"

type Role int

const (
	Follower = iota
	Candidate
	Leader
)

func (s *RaftServer) becomeCandidate() {
	s.mu.Lock()

	s.logger.Info("Become candidate and start election", zap.Int64("node_id", s.id))

	s.state = Candidate
	s.curTerm++
	s.votedFor = &s.id

	s.mu.Unlock()

	s.resetElectionTimer()
	s.startElection()
}

func (s *RaftServer) becomeLeader() {
	s.mu.Lock()

	s.logger.Info("Become leader", zap.Int64("node_id", s.id))

	s.state = Leader
	s.curLeader = s.id
	s.nextIndex = make(map[PeerAddr]int64)
	s.matchIndex = make(map[int64]int)

	if s.electionTimer != nil {
		s.electionTimer.Stop()
	}

	s.mu.Unlock()

	s.resetHeartbeatTimer()
	go s.sendHeartbeats()
}

func (s *RaftServer) becomeFollower(term int64, leaderID int64) {
	s.mu.Lock()

	s.logger.Info("Become follower", zap.Int64("term", term), zap.Int64("node_id", s.id))

	s.state = Follower
	s.curTerm = term
	s.votedFor = nil
	s.curLeader = leaderID

	s.mu.Unlock()

	s.resetElectionTimer()
}
