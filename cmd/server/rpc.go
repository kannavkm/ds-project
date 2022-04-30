package main

import "sync"
z
// server accepts rpc responses from the multiple worker nodes
// and Also sends rpc requests to appropriate worker nodes

type groupInfo struct {

}

type grpcServer struct {
	mutex sync.Mutex
}
