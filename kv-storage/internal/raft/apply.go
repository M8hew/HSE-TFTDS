package raft

import "go.uber.org/zap"

const (
	SET = "set"
	CAS = "cas"
	UPD = "upd"
	DEL = "del"
)

func NewSetEntry(key, value string) LogEntry {
	return LogEntry{
		Command: SET,
		Key:     key,
		Value:   &value,
	}
}

func NewCASEntry(key, oldValue, newValue string) LogEntry {
	return LogEntry{
		Command:  CAS,
		Key:      key,
		Value:    &oldValue,
		OldValue: &newValue,
	}
}

func NewUpdateEntry(key, value string) LogEntry {
	return LogEntry{
		Command: UPD,
		Key:     key,
		Value:   &value,
	}
}

func NewDeleteEntry(key string) LogEntry {
	return LogEntry{
		Command: DEL,
		Key:     key,
	}
}

func (s *RaftServer) apply(entry LogEntry) {
	switch entry.Command {
	case SET:
		s.logger.Debug("Applying SET command", zap.Int64("node_id", s.id))
		s.storage.Set(entry.Key, *entry.Value)
	case CAS:
		s.logger.Debug("Applying CAS command", zap.Int64("node_id", s.id))
		s.storage.CAS(entry.Key, *entry.OldValue, *entry.Value)
	case UPD:
		s.logger.Debug("Applying UPD command", zap.Int64("node_id", s.id))
		s.storage.Update(entry.Key, *entry.Value)
	case DEL:
		s.logger.Debug("Applying DEL command", zap.Int64("node_id", s.id))
		s.storage.Del(entry.Key)
	default:
		panic("unknown command: " + entry.Command)
	}
}
