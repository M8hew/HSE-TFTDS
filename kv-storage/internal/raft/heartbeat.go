package raft

import (
	"go.uber.org/zap"

	pb "kvstorage/api/proto"
)

func (s *RaftServer) sendHeartbeats() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != Leader {
		return
	}

	for _, peer := range s.peers {
		go func(peer PeerAddr) {
			for {
				s.mu.Lock()
				defer s.mu.Unlock()

				var (
					prevLogTerm int64
					nextInd     int64
					ok          bool
				)

				if nextInd, ok = s.nextIndex[peer]; !ok {
					s.nextIndex[peer] = 0
				}

				if nextInd > 0 {
					prevLogTerm = s.log[nextInd-1].Term
				}

				req := &pb.AppendEntriesRequest{
					Term:         s.curTerm,
					LeaderID:     s.id,
					LeaderCommit: s.commitIndex,
					PrevLogIndex: nextInd - 1,
					PrevLogTerm:  prevLogTerm,
					Entries:      toPbEntrySlice(s.log[nextInd:]),
				}

				s.mu.Unlock()

				res, err := sendAppendEntries(peer, req)
				if err != nil {
					s.logger.Error("sendHeartbeats error", zap.String("peer_id", string(peer)), zap.Error(err), zap.Int64("node_id", s.id))
				}

				if nextInd == 0 {
					return
				}

				s.mu.Lock()

				if !res.Success {
					s.logger.Info("Replica sync broken", zap.String("peer_id", string(peer)), zap.Int64("leader", s.id), zap.Int64("node_id", s.id))
					s.nextIndex[peer] = nextInd - 1
					return
				}

				s.nextIndex[peer] = nextInd + int64(len(req.Entries))

				count := 0
				for _, ind := range s.nextIndex {
					if ind > s.commitIndex+1 {
						count++
					}
				}

				if count > len(s.peers)/2 {
					s.commitIndex++
					s.apply(s.log[s.commitIndex])
				}
			}
		}(peer)
	}

	s.resetHeartbeatTimer()
}
