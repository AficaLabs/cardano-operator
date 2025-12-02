/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cardano

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Executor implements the Client interface using cardano-cli
type Executor struct {
	config *ClientConfig
}

// NewExecutor creates a new cardano-cli executor
func NewExecutor(config *ClientConfig) (*Executor, error) {
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 30
	}
	if config.CLIPath == "" {
		config.CLIPath = "cardano-cli"
	}
	if config.NetworkMagic == 0 {
		switch config.Network {
		case NetworkMainnet:
			config.NetworkMagic = 764824073
		case NetworkPreprod:
			config.NetworkMagic = 1
		case NetworkPreview:
			config.NetworkMagic = 2
		}
	}
	return &Executor{config: config}, nil
}

// runCLI executes a cardano-cli command
func (e *Executor) runCLI(ctx context.Context, args ...string) ([]byte, error) {
	timeout := time.Duration(e.config.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.config.CLIPath, args...)

	if e.config.SocketPath != "" {
		cmd.Env = append(os.Environ(), "CARDANO_NODE_SOCKET_PATH="+e.config.SocketPath)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cardano-cli error: %v, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// networkArgs returns the network arguments for cardano-cli
func (e *Executor) networkArgs() []string {
	if e.config.Network == NetworkMainnet {
		return []string{"--mainnet"}
	}
	return []string{"--testnet-magic", strconv.FormatInt(e.config.NetworkMagic, 10)}
}

// QueryTip queries the current chain tip
func (e *Executor) QueryTip(ctx context.Context) (*Tip, error) {
	args := append([]string{"query", "tip"}, e.networkArgs()...)

	output, err := e.runCLI(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tip: %w", err)
	}

	var tip Tip
	if err := json.Unmarshal(output, &tip); err != nil {
		return nil, fmt.Errorf("failed to parse tip response: %w", err)
	}

	return &tip, nil
}

// QueryUTxO queries UTxOs for a given address
func (e *Executor) QueryUTxO(ctx context.Context, address string) ([]UTxO, error) {
	args := append([]string{"query", "utxo", "--address", address, "--out-file", "/dev/stdout"}, e.networkArgs()...)

	output, err := e.runCLI(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query utxo: %w", err)
	}

	// Parse the UTxO output (JSON format)
	var utxoMap map[string]interface{}
	if err := json.Unmarshal(output, &utxoMap); err != nil {
		return nil, fmt.Errorf("failed to parse utxo response: %w", err)
	}

	utxos := make([]UTxO, 0, len(utxoMap))
	for key, val := range utxoMap {
		parts := strings.Split(key, "#")
		if len(parts) != 2 {
			continue
		}
		txIndex, _ := strconv.Atoi(parts[1])

		utxoData, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		utxo := UTxO{
			TxHash:  parts[0],
			TxIndex: txIndex,
			Address: address,
		}

		// Parse value
		if valueData, ok := utxoData["value"].(map[string]interface{}); ok {
			if lovelace, ok := valueData["lovelace"].(float64); ok {
				utxo.Value.Lovelace = int64(lovelace)
			}
		}

		utxos = append(utxos, utxo)
	}

	return utxos, nil
}

// QueryProtocolParameters queries the current protocol parameters
func (e *Executor) QueryProtocolParameters(ctx context.Context) (*ProtocolParameters, error) {
	args := append([]string{"query", "protocol-parameters", "--out-file", "/dev/stdout"}, e.networkArgs()...)

	output, err := e.runCLI(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query protocol parameters: %w", err)
	}

	var params ProtocolParameters
	if err := json.Unmarshal(output, &params); err != nil {
		return nil, fmt.Errorf("failed to parse protocol parameters: %w", err)
	}

	return &params, nil
}

// QueryPoolParams queries the parameters of a registered pool
func (e *Executor) QueryPoolParams(ctx context.Context, poolID string) (*PoolParams, error) {
	args := append([]string{"query", "pool-params", "--stake-pool-id", poolID, "--out-file", "/dev/stdout"}, e.networkArgs()...)

	output, err := e.runCLI(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pool params: %w", err)
	}

	var params PoolParams
	if err := json.Unmarshal(output, &params); err != nil {
		return nil, fmt.Errorf("failed to parse pool params: %w", err)
	}
	params.PoolID = poolID

	return &params, nil
}

// QueryStakePoolID derives the pool ID from a cold verification key
func (e *Executor) QueryStakePoolID(ctx context.Context, coldVKeyPath string) (string, error) {
	args := []string{"stake-pool", "id", "--cold-verification-key-file", coldVKeyPath}

	output, err := e.runCLI(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to query pool id: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GenerateColdKeys generates cold key pair
func (e *Executor) GenerateColdKeys(ctx context.Context) (*KeyPair, error) {
	return e.generateColdKeyPair(ctx)
}

// GenerateVRFKeys generates VRF key pair
func (e *Executor) GenerateVRFKeys(ctx context.Context) (*KeyPair, error) {
	return e.generateKeyPair(ctx, "node", "key-gen-VRF", "--verification-key-file", "--signing-key-file")
}

// GenerateKESKeys generates KES key pair
func (e *Executor) GenerateKESKeys(ctx context.Context) (*KeyPair, error) {
	return e.generateKeyPair(ctx, "node", "key-gen-KES", "--verification-key-file", "--signing-key-file")
}

// GeneratePaymentKeys generates payment key pair
func (e *Executor) GeneratePaymentKeys(ctx context.Context) (*KeyPair, error) {
	return e.generateKeyPair(ctx, "latest", "address", "key-gen", "--verification-key-file", "--signing-key-file")
}

// GenerateStakeKeys generates stake key pair
func (e *Executor) GenerateStakeKeys(ctx context.Context) (*KeyPair, error) {
	return e.generateKeyPair(ctx, "latest", "stake-address", "key-gen", "--verification-key-file", "--signing-key-file")
}

// generateColdKeyPair generates cold keys (node key-gen), which requires additional counter file in CLI v10+
func (e *Executor) generateColdKeyPair(ctx context.Context) (*KeyPair, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-keys-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	vkeyPath := filepath.Join(tmpDir, "cold.vkey")
	skeyPath := filepath.Join(tmpDir, "cold.skey")
	counterPath := filepath.Join(tmpDir, "cold.counter")

	args := []string{
		"node", "key-gen",
		"--cold-verification-key-file", vkeyPath,
		"--cold-signing-key-file", skeyPath,
		"--operational-certificate-issue-counter-file", counterPath,
	}

	if _, err := e.runCLI(ctx, args...); err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	vkey, err := os.ReadFile(vkeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read verification key: %w", err)
	}

	skey, err := os.ReadFile(skeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signing key: %w", err)
	}

	return &KeyPair{
		VerificationKey: vkey,
		SigningKey:      skey,
	}, nil
}

// generateKeyPair is a helper for generating key pairs
// Supports variable number of command segments (e.g., "node", "key-gen-VRF" or "latest", "address", "key-gen")
func (e *Executor) generateKeyPair(ctx context.Context, cmdParts ...string) (*KeyPair, error) {
	// Last two parts are vkeyFlag and skeyFlag
	if len(cmdParts) < 4 {
		return nil, fmt.Errorf("insufficient arguments for generateKeyPair")
	}

	skeyFlag := cmdParts[len(cmdParts)-1]
	vkeyFlag := cmdParts[len(cmdParts)-2]
	commandParts := cmdParts[:len(cmdParts)-2]

	tmpDir, err := os.MkdirTemp("", "cardano-keys-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	vkeyPath := filepath.Join(tmpDir, "key.vkey")
	skeyPath := filepath.Join(tmpDir, "key.skey")

	args := append(commandParts, vkeyFlag, vkeyPath, skeyFlag, skeyPath)

	if _, err := e.runCLI(ctx, args...); err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	vkey, err := os.ReadFile(vkeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read verification key: %w", err)
	}

	skey, err := os.ReadFile(skeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signing key: %w", err)
	}

	return &KeyPair{
		VerificationKey: vkey,
		SigningKey:      skey,
	}, nil
}

// CreateOperationalCertificate creates a KES operational certificate
func (e *Executor) CreateOperationalCertificate(ctx context.Context, kesSKey []byte, coldSKey []byte, counter int64, kesPeriod int64) (*OperationalCertificate, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-opcert-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	kesVKeyPath := filepath.Join(tmpDir, "kes.vkey")
	coldSKeyPath := filepath.Join(tmpDir, "cold.skey")
	counterPath := filepath.Join(tmpDir, "counter")
	certPath := filepath.Join(tmpDir, "node.cert")

	// We need to derive the KES verification key from signing key first
	kesSKeyPath := filepath.Join(tmpDir, "kes.skey")
	if err := os.WriteFile(kesSKeyPath, kesSKey, 0600); err != nil {
		return nil, fmt.Errorf("failed to write KES signing key: %w", err)
	}

	if err := os.WriteFile(coldSKeyPath, coldSKey, 0600); err != nil {
		return nil, fmt.Errorf("failed to write cold signing key: %w", err)
	}

	// Create counter file
	counterJSON := fmt.Sprintf(`{"type":"NodeOperationalCertificateIssueCounter","description":"","cborHex":"%016x"}`, counter)
	if err := os.WriteFile(counterPath, []byte(counterJSON), 0600); err != nil {
		return nil, fmt.Errorf("failed to write counter: %w", err)
	}

	// Get KES vkey from skey
	args := []string{"key", "verification-key", "--signing-key-file", kesSKeyPath, "--verification-key-file", kesVKeyPath}
	if _, err := e.runCLI(ctx, args...); err != nil {
		return nil, fmt.Errorf("failed to derive KES verification key: %w", err)
	}

	// Issue the certificate
	args = []string{
		"node", "issue-op-cert",
		"--kes-verification-key-file", kesVKeyPath,
		"--cold-signing-key-file", coldSKeyPath,
		"--operational-certificate-issue-counter-file", counterPath,
		"--kes-period", strconv.FormatInt(kesPeriod, 10),
		"--out-file", certPath,
	}

	if _, err := e.runCLI(ctx, args...); err != nil {
		return nil, fmt.Errorf("failed to issue operational certificate: %w", err)
	}

	cert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	// Calculate expiry based on protocol parameters
	params, err := e.QueryProtocolParameters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query protocol parameters: %w", err)
	}

	tip, err := e.QueryTip(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query tip: %w", err)
	}

	expiryEpoch := tip.Epoch + params.MaxKESEvolutions

	return &OperationalCertificate{
		Certificate: cert,
		Counter:     counter + 1,
		KESPeriod:   kesPeriod,
		ExpiryEpoch: expiryEpoch,
	}, nil
}

// CreatePoolRegistrationCertificate creates a pool registration certificate
func (e *Executor) CreatePoolRegistrationCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
	return e.createPoolCertificate(ctx, params, coldVKey, false)
}

// CreatePoolUpdateCertificate creates a pool update certificate
func (e *Executor) CreatePoolUpdateCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
	return e.createPoolCertificate(ctx, params, coldVKey, false)
}

func (e *Executor) createPoolCertificate(ctx context.Context, params *PoolParams, coldVKey []byte, _ bool) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-poolcert-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	coldVKeyPath := filepath.Join(tmpDir, "cold.vkey")
	vrfVKeyPath := filepath.Join(tmpDir, "vrf.vkey")
	certPath := filepath.Join(tmpDir, "pool.cert")

	if err := os.WriteFile(coldVKeyPath, coldVKey, 0600); err != nil {
		return nil, fmt.Errorf("failed to write cold vkey: %w", err)
	}

	// Create VRF vkey file from hash (simplified - in real impl would need actual key)
	vrfVKeyJSON := fmt.Sprintf(`{"type":"VrfVerificationKey_PraosVRF","description":"","cborHex":"%s"}`, params.VRFKeyHash)
	if err := os.WriteFile(vrfVKeyPath, []byte(vrfVKeyJSON), 0600); err != nil {
		return nil, fmt.Errorf("failed to write vrf vkey: %w", err)
	}

	args := []string{
		"stake-pool", "registration-certificate",
		"--cold-verification-key-file", coldVKeyPath,
		"--vrf-verification-key-file", vrfVKeyPath,
		"--pool-pledge", strconv.FormatInt(params.Pledge, 10),
		"--pool-cost", strconv.FormatInt(params.Cost, 10),
		"--pool-margin", fmt.Sprintf("%.4f", params.Margin),
		"--pool-reward-account-verification-key-file", coldVKeyPath,
		"--pool-owner-stake-verification-key-file", coldVKeyPath,
		"--out-file", certPath,
	}
	args = append(args, e.networkArgs()...)

	// Add relay configurations
	for _, relay := range params.Relays {
		switch relay.Type {
		case "dns":
			args = append(args, "--single-host-pool-relay", relay.Hostname)
			args = append(args, "--pool-relay-port", strconv.FormatInt(int64(relay.Port), 10))
		case "ip":
			if relay.IPv4 != "" {
				args = append(args, "--pool-relay-ipv4", relay.IPv4)
			}
			if relay.IPv6 != "" {
				args = append(args, "--pool-relay-ipv6", relay.IPv6)
			}
			args = append(args, "--pool-relay-port", strconv.FormatInt(int64(relay.Port), 10))
		}
	}

	// Add metadata if present
	if params.MetadataURL != "" {
		args = append(args, "--metadata-url", params.MetadataURL)
		if params.MetadataHash != "" {
			args = append(args, "--metadata-hash", params.MetadataHash)
		}
	}

	if _, err := e.runCLI(ctx, args...); err != nil {
		return nil, fmt.Errorf("failed to create pool certificate: %w", err)
	}

	return os.ReadFile(certPath)
}

// CreatePoolRetirementCertificate creates a pool retirement certificate
func (e *Executor) CreatePoolRetirementCertificate(ctx context.Context, poolID string, retirementEpoch int64, coldVKey []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-retire-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	coldVKeyPath := filepath.Join(tmpDir, "cold.vkey")
	certPath := filepath.Join(tmpDir, "retire.cert")

	if err := os.WriteFile(coldVKeyPath, coldVKey, 0600); err != nil {
		return nil, fmt.Errorf("failed to write cold vkey: %w", err)
	}

	args := []string{
		"stake-pool", "deregistration-certificate",
		"--cold-verification-key-file", coldVKeyPath,
		"--epoch", strconv.FormatInt(retirementEpoch, 10),
		"--out-file", certPath,
	}

	if _, err := e.runCLI(ctx, args...); err != nil {
		return nil, fmt.Errorf("failed to create retirement certificate: %w", err)
	}

	return os.ReadFile(certPath)
}

// BuildTx builds a transaction
func (e *Executor) BuildTx(ctx context.Context, opts *TxBuildOptions) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-tx-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	txBodyPath := filepath.Join(tmpDir, "tx.raw")

	args := []string{"transaction", "build"}
	args = append(args, e.networkArgs()...)

	// Add inputs
	for _, input := range opts.Inputs {
		args = append(args, "--tx-in", fmt.Sprintf("%s#%d", input.TxHash, input.TxIndex))
	}

	// Add outputs
	for _, output := range opts.Outputs {
		args = append(args, "--tx-out", fmt.Sprintf("%s+%d", output.Address, output.Value.Lovelace))
	}

	// Add certificates
	for i, cert := range opts.Certificates {
		certPath := filepath.Join(tmpDir, fmt.Sprintf("cert%d.cert", i))
		if err := os.WriteFile(certPath, cert, 0600); err != nil {
			return nil, fmt.Errorf("failed to write certificate: %w", err)
		}
		args = append(args, "--certificate-file", certPath)
	}

	// Add change address
	args = append(args, "--change-address", opts.ChangeAddress)

	// Add output file
	args = append(args, "--out-file", txBodyPath)

	if _, err := e.runCLI(ctx, args...); err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	return os.ReadFile(txBodyPath)
}

// SignTx signs a transaction
func (e *Executor) SignTx(ctx context.Context, txBody []byte, signingKeys ...[]byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-sign-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	txBodyPath := filepath.Join(tmpDir, "tx.raw")
	txSignedPath := filepath.Join(tmpDir, "tx.signed")

	if err := os.WriteFile(txBodyPath, txBody, 0600); err != nil {
		return nil, fmt.Errorf("failed to write tx body: %w", err)
	}

	args := []string{"transaction", "sign", "--tx-body-file", txBodyPath}
	args = append(args, e.networkArgs()...)

	// Add signing keys
	for i, key := range signingKeys {
		keyPath := filepath.Join(tmpDir, fmt.Sprintf("key%d.skey", i))
		if err := os.WriteFile(keyPath, key, 0600); err != nil {
			return nil, fmt.Errorf("failed to write signing key: %w", err)
		}
		args = append(args, "--signing-key-file", keyPath)
	}

	args = append(args, "--out-file", txSignedPath)

	if _, err := e.runCLI(ctx, args...); err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return os.ReadFile(txSignedPath)
}

// SubmitTx submits a signed transaction
func (e *Executor) SubmitTx(ctx context.Context, signedTx []byte) (*TransactionSubmitResult, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-submit-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	txPath := filepath.Join(tmpDir, "tx.signed")
	if err := os.WriteFile(txPath, signedTx, 0600); err != nil {
		return nil, fmt.Errorf("failed to write signed tx: %w", err)
	}

	args := []string{"transaction", "submit", "--tx-file", txPath}
	args = append(args, e.networkArgs()...)

	output, err := e.runCLI(ctx, args...)
	if err != nil {
		return &TransactionSubmitResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Get tx hash
	hashArgs := []string{"transaction", "txid", "--tx-file", txPath}
	hashOutput, err := e.runCLI(ctx, hashArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx hash: %w", err)
	}

	return &TransactionSubmitResult{
		TxHash:  strings.TrimSpace(string(hashOutput)),
		Success: true,
		Error:   string(output),
	}, nil
}

// BuildAddress builds a payment address from keys
func (e *Executor) BuildAddress(ctx context.Context, paymentVKey []byte, stakeVKey []byte) (string, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-addr-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	paymentVKeyPath := filepath.Join(tmpDir, "payment.vkey")
	stakeVKeyPath := filepath.Join(tmpDir, "stake.vkey")

	if err := os.WriteFile(paymentVKeyPath, paymentVKey, 0600); err != nil {
		return "", fmt.Errorf("failed to write payment vkey: %w", err)
	}

	if err := os.WriteFile(stakeVKeyPath, stakeVKey, 0600); err != nil {
		return "", fmt.Errorf("failed to write stake vkey: %w", err)
	}

	args := []string{
		"address", "build",
		"--payment-verification-key-file", paymentVKeyPath,
		"--stake-verification-key-file", stakeVKeyPath,
	}
	args = append(args, e.networkArgs()...)

	output, err := e.runCLI(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to build address: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// BuildStakeAddress builds a stake address from a stake verification key
func (e *Executor) BuildStakeAddress(ctx context.Context, stakeVKey []byte) (string, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-stake-addr-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	stakeVKeyPath := filepath.Join(tmpDir, "stake.vkey")
	if err := os.WriteFile(stakeVKeyPath, stakeVKey, 0600); err != nil {
		return "", fmt.Errorf("failed to write stake vkey: %w", err)
	}

	args := []string{
		"latest", "stake-address", "build",
		"--stake-verification-key-file", stakeVKeyPath,
	}
	args = append(args, e.networkArgs()...)

	output, err := e.runCLI(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to build stake address: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CalculateMinFee calculates the minimum fee for a transaction
func (e *Executor) CalculateMinFee(ctx context.Context, txBody []byte, witnessCount int) (int64, error) {
	tmpDir, err := os.MkdirTemp("", "cardano-fee-*")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	txBodyPath := filepath.Join(tmpDir, "tx.raw")
	protocolPath := filepath.Join(tmpDir, "protocol.json")

	if err := os.WriteFile(txBodyPath, txBody, 0600); err != nil {
		return 0, fmt.Errorf("failed to write tx body: %w", err)
	}

	// Get protocol parameters
	params, err := e.QueryProtocolParameters(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query protocol parameters: %w", err)
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal protocol parameters: %w", err)
	}

	if err := os.WriteFile(protocolPath, paramsJSON, 0600); err != nil {
		return 0, fmt.Errorf("failed to write protocol parameters: %w", err)
	}

	args := []string{
		"transaction", "calculate-min-fee",
		"--tx-body-file", txBodyPath,
		"--protocol-params-file", protocolPath,
		"--tx-in-count", "1",
		"--tx-out-count", "1",
		"--witness-count", strconv.Itoa(witnessCount),
	}

	output, err := e.runCLI(ctx, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate min fee: %w", err)
	}

	// Parse "X Lovelace" format
	parts := strings.Fields(string(output))
	if len(parts) < 1 {
		return 0, fmt.Errorf("unexpected fee output format: %s", output)
	}

	fee, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse fee: %w", err)
	}

	return fee, nil
}

// GetCurrentKESPeriod calculates the current KES period
func (e *Executor) GetCurrentKESPeriod(ctx context.Context) (int64, error) {
	tip, err := e.QueryTip(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query tip: %w", err)
	}

	params, err := e.QueryProtocolParameters(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query protocol parameters: %w", err)
	}

	if params.SlotsPerKESPeriod == 0 {
		return 0, fmt.Errorf("slotsPerKESPeriod is 0")
	}

	kesPeriod := tip.Slot / params.SlotsPerKESPeriod
	return kesPeriod, nil
}

// Ensure Executor implements Client
var _ Client = (*Executor)(nil)
