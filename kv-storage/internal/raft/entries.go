package raft

import (
	"context"

	"go.uber.org/zap"

	pb "kvstorage/api/proto"
)

type LogEntry struct {
	Term     int64
	Command  string
	Key      string
	Value    *string
	OldValue *string
}

func toPbEntry(entry *LogEntry) *pb.LogEntry {
	return &pb.LogEntry{
		Term:     entry.Term,
		Command:  entry.Command,
		Key:      entry.Key,
		Value:    entry.Value,
		OldValue: entry.OldValue,
	}
}

func fromPbEntry(entry *pb.LogEntry) LogEntry {
	return LogEntry{
		Term:     entry.Term,
		Command:  entry.Command,
		Key:      entry.Key,
		Value:    entry.Value,
		OldValue: entry.Value,
	}
}

func toPbEntrySlice(entries []LogEntry) []*pb.LogEntry {
	res := make([]*pb.LogEntry, 0, len(entries))
	for _, entry := range entries {
		res = append(res, toPbEntry(&entry))
	}
	return res
}

// grpc handler
func (s *RaftServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	s.logger.Info("AppendEntries handler", zap.Int64("node_id", s.id))

	failedRequest := pb.AppendEntriesResponse{
		Term:    s.curTerm,
		Success: false,
	}

	if req.Term < s.curTerm {
		return &failedRequest, nil
	}

	s.becomeFollower(req.Term, req.LeaderID)

	if req.PrevLogIndex >= 0 {
		if req.PrevLogIndex >= int64(len(s.log)) {
			return &failedRequest, nil
		}
		if req.PrevLogTerm != s.log[req.PrevLogIndex].Term {
			return &failedRequest, nil
		}
	}

	s.mu.Lock()

	for i, entry := range req.Entries {
		ind := req.PrevLogIndex + int64(i) + 1
		if ind < int64(len(s.log)) {
			if req.Term == s.log[ind].Term {
				continue
			}
			s.log = s.log[:ind]
		}
		s.log = append(s.log, fromPbEntry(entry))
	}

	s.mu.Unlock()

	if req.LeaderCommit > s.commitIndex {
		s.mu.Lock()

		for i := s.commitIndex; i <= req.LeaderCommit; i++ {
			s.apply(s.log[i])
		}
		s.commitIndex = req.LeaderCommit

		s.mu.Unlock()
	}

	return &pb.AppendEntriesResponse{
		Term:    s.curTerm,
		Success: true,
	}, nil
}
