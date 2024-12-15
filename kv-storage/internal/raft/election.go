package raft

import (
	"context"

	"go.uber.org/zap"

	pb "kvstorage/api/proto"
)

func (s *RaftServer) startElection() {
	votes := 1
	for _, peer := range s.peers {
		lastLogInd := int64(len(s.log) - 1)
		lastLogTerm := int64(0)
		if len(s.log) > 0 {
			lastLogTerm = s.log[lastLogInd].Term
		}

		go func(peer PeerAddr) {
			req := &pb.RequestVoteRequest{
				Term:         s.curTerm,
				CandidateID:  s.id,
				LastLogIndex: lastLogInd,
				LastLogTerm:  lastLogTerm,
			}

			resp, err := sendRequestVote(peer, req)
			if err != nil {
				s.logger.Error("startElection error", zap.String("peer_id", string(peer)), zap.Error(err))
				return
			}

			if resp.VoteGranted {
				s.mu.Lock()
				defer s.mu.Unlock()

				votes++
				if votes > len(s.peers)/2 && s.state == Candidate {
					s.becomeLeader()
				}
			}
		}(peer)
	}
}

// grpc request handler
func (s *RaftServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	s.logger.Info("RequestVote handler")

	failedRequest := pb.RequestVoteResponse{
		Term:        s.curTerm,
		VoteGranted: false,
	}

	if req.Term < s.curTerm {
		return &failedRequest, nil
	}

	if s.votedFor != nil && req.Term == s.curTerm && *s.votedFor != req.CandidateID {
		return &failedRequest, nil
	}

	lastLogInd := len(s.log) - 1

	var lastLogTerm int64
	if len(s.log) > 0 {
		lastLogTerm = s.log[lastLogInd].Term
	}

	if req.LastLogTerm < lastLogTerm {
		return &failedRequest, nil
	}

	if req.LastLogTerm == lastLogTerm && req.LastLogIndex < int64(lastLogInd) {
		return &failedRequest, nil
	}

	s.curTerm = req.Term
	s.votedFor = &req.CandidateID
	s.resetElectionTimer()

	return &pb.RequestVoteResponse{
		Term:        s.curTerm,
		VoteGranted: true,
	}, nil
}
