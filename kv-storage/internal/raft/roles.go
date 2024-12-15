package raft

import "go.uber.org/zap"

type Role int

const (
	_ = iota
	Leader
	Follower
	Candidate
)

func (s *RaftServer) becomeCandidate() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Become candidate and start election")

	s.state = Candidate
	s.curTerm++
	s.votedFor = &s.id

	s.resetElectionTimer()
	s.startElection()
}

func (s *RaftServer) becomeLeader() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Become leader")

	s.state = Leader
	s.curLeader = s.id
	s.nextIndex = make(map[PeerAddr]int64)
	s.matchIndex = make(map[int64]int)

	if s.electionTimer != nil {
		s.electionTimer.Stop()
	}

	s.resetHeartbeatTimer()
	s.sendHeartbeats()
}

func (s *RaftServer) becomeFollower(term int64, leaderID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Become follower", zap.Int64("term", term))

	s.state = Follower
	s.curTerm = term
	s.votedFor = nil
	s.curLeader = leaderID

	s.resetElectionTimer()
}
