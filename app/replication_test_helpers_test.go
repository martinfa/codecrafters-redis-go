package main

func resetReplicationStateForTest() {
	replicasMutex.Lock()
	defer replicasMutex.Unlock()

	replicas = nil
	masterReplicationOffset = 0
}
