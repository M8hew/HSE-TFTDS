package raft

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "kvstorage/api/proto"
)

func sendRequestVote(peerAddr PeerAddr, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	conn, err := grpc.NewClient(string(peerAddr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return client.RequestVote(ctx, req)

}

func sendAppendEntries(peerAddr PeerAddr, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	conn, err := grpc.NewClient(string(peerAddr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return client.AppendEntries(ctx, req)
}
