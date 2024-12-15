package raft

import "time"

func (s *RaftServer) resetElectionTimer() {
	if s.electionTimer != nil {
		s.electionTimer.Stop()
	}

	s.electionTimer = time.AfterFunc(s.electionTimeout, func() {
		s.becomeCandidate()
	})
}

func (s *RaftServer) resetHeartbeatTimer() {
	if s.heartbeatTimer != nil {
		s.heartbeatTimer.Stop()
	}

	s.heartbeatTimer = time.AfterFunc(s.heartbeatTimeout, func() {
		s.sendHeartbeats()
	})
}
