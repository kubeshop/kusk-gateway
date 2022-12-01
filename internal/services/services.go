package services

// AuthServiceHostPort
// Previously was `kusk-gateway-auth-service`.
func AuthServiceHostPort() (string, int) {
	port := 19000
	return "kusk-gateway-manager.kusk-system.svc.cluster.local", port
}

// ValidatorHostPort
// Previously was `kusk-gateway-validator-service`.
func ValidatorHostPort() (string, int) {
	port := 17000
	return "kusk-gateway-manager.kusk-system.svc.cluster.local", port
}
