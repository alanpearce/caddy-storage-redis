package storageredis

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"go.uber.org/zap"
)

func (rs *RedisStorage) ProvisionCertMagic(ctx context.Context, log *zap.Logger) error {
	rs.logger = log

	err := rs.finalizeConfiguration(ctx)
	if err != nil {
		rs.logger.Info("Provision Redis storage", zap.String("client_type", rs.ClientType), zap.Strings("address", rs.Address))
	}

	return err
}

func (rs *RedisStorage) finalizeConfiguration(ctx context.Context) error {

	for idx, v := range rs.Address {
		host, port, err := net.SplitHostPort(v)
		if err != nil {
			return fmt.Errorf("invalid address: %s", v)
		}
		rs.Address[idx] = net.JoinHostPort(host, port)
	}
	for idx, v := range rs.Host {
		addr := net.ParseIP(v)
		_, err := net.LookupHost(v)
		if addr == nil && err != nil {
			return fmt.Errorf("invalid host value: %s", v)
		}
		rs.Host[idx] = v
	}
	for idx, v := range rs.Port {
		_, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid port value: %s", v)
		}
		rs.Port[idx] = v
	}
	if rs.Timeout != "" {
		timeParse, err := strconv.Atoi(rs.Timeout)
		if err != nil || timeParse < 0 {
			return fmt.Errorf("invalid timeout value: %s", rs.Timeout)
		}
	}

	if len(rs.EncryptionKey) > 0 {
		// Encryption_key length must be at least 32 characters
		if len(rs.EncryptionKey) < 32 {
			return fmt.Errorf("invalid length for 'encryption_key', must contain at least 32 bytes: %s", rs.EncryptionKey)
		}
		// Truncate keys that are too long
		if len(rs.EncryptionKey) > 32 {
			rs.EncryptionKey = rs.EncryptionKey[:32]
		}
	}

	// Construct Address from Host and Port if not explicitly provided
	if len(rs.Address) == 0 {

		var maxAddrs int
		var host, port string

		maxHosts := len(rs.Host)
		maxPorts := len(rs.Port)

		// Determine max number of addresses
		if maxHosts > maxPorts {
			maxAddrs = maxHosts
		} else {
			maxAddrs = maxPorts
		}

		for i := 0; i < maxAddrs; i++ {
			if i < maxHosts {
				host = rs.Host[i]
			}
			if i < maxPorts {
				port = rs.Port[i]
			}
			rs.Address = append(rs.Address, net.JoinHostPort(host, port))
		}
		// Clear host and port values
		rs.Host = []string{}
		rs.Port = []string{}
	}

	return rs.initRedisClient(ctx)
}

func (rs *RedisStorage) Cleanup() error {
	// Close the Redis connection
	if rs.client != nil {
		rs.client.Close()
	}

	return nil
}
