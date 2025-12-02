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
	"testing"
)

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name        string
		config      *ClientConfig
		wantTimeout int
		wantCLIPath string
		wantMagic   int64
		wantErr     bool
	}{
		{
			name: "default values for preprod",
			config: &ClientConfig{
				Network: NetworkPreprod,
			},
			wantTimeout: 30,
			wantCLIPath: "cardano-cli",
			wantMagic:   1,
			wantErr:     false,
		},
		{
			name: "default values for mainnet",
			config: &ClientConfig{
				Network: NetworkMainnet,
			},
			wantTimeout: 30,
			wantCLIPath: "cardano-cli",
			wantMagic:   764824073,
			wantErr:     false,
		},
		{
			name: "default values for preview",
			config: &ClientConfig{
				Network: NetworkPreview,
			},
			wantTimeout: 30,
			wantCLIPath: "cardano-cli",
			wantMagic:   2,
			wantErr:     false,
		},
		{
			name: "custom timeout",
			config: &ClientConfig{
				Network:        NetworkPreprod,
				TimeoutSeconds: 60,
			},
			wantTimeout: 60,
			wantCLIPath: "cardano-cli",
			wantMagic:   1,
			wantErr:     false,
		},
		{
			name: "custom CLI path",
			config: &ClientConfig{
				Network: NetworkPreprod,
				CLIPath: "/usr/local/bin/cardano-cli",
			},
			wantTimeout: 30,
			wantCLIPath: "/usr/local/bin/cardano-cli",
			wantMagic:   1,
			wantErr:     false,
		},
		{
			name: "custom network magic",
			config: &ClientConfig{
				Network:      NetworkPreprod,
				NetworkMagic: 12345,
			},
			wantTimeout: 30,
			wantCLIPath: "cardano-cli",
			wantMagic:   12345,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewExecutor(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewExecutor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if executor.config.TimeoutSeconds != tt.wantTimeout {
				t.Errorf("TimeoutSeconds = %v, want %v", executor.config.TimeoutSeconds, tt.wantTimeout)
			}
			if executor.config.CLIPath != tt.wantCLIPath {
				t.Errorf("CLIPath = %v, want %v", executor.config.CLIPath, tt.wantCLIPath)
			}
			if executor.config.NetworkMagic != tt.wantMagic {
				t.Errorf("NetworkMagic = %v, want %v", executor.config.NetworkMagic, tt.wantMagic)
			}
		})
	}
}

func TestExecutorNetworkArgs(t *testing.T) {
	tests := []struct {
		name    string
		network Network
		magic   int64
		want    []string
	}{
		{
			name:    "mainnet",
			network: NetworkMainnet,
			magic:   764824073,
			want:    []string{"--mainnet"},
		},
		{
			name:    "preprod",
			network: NetworkPreprod,
			magic:   1,
			want:    []string{"--testnet-magic", "1"},
		},
		{
			name:    "preview",
			network: NetworkPreview,
			magic:   2,
			want:    []string{"--testnet-magic", "2"},
		},
		{
			name:    "custom magic",
			network: NetworkPreprod,
			magic:   12345,
			want:    []string{"--testnet-magic", "12345"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &Executor{
				config: &ClientConfig{
					Network:      tt.network,
					NetworkMagic: tt.magic,
				},
			}

			args := executor.networkArgs()

			if len(args) != len(tt.want) {
				t.Errorf("networkArgs() = %v, want %v", args, tt.want)
				return
			}

			for i, arg := range args {
				if arg != tt.want[i] {
					t.Errorf("networkArgs()[%d] = %v, want %v", i, arg, tt.want[i])
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	config := &ClientConfig{
		Network: NetworkPreprod,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
		return
	}

	if client == nil {
		t.Error("NewClient() returned nil client")
	}

	// Verify it returns an Executor
	_, ok := client.(*Executor)
	if !ok {
		t.Error("NewClient() did not return an *Executor")
	}
}

func TestNetworkConstants(t *testing.T) {
	if NetworkMainnet != "mainnet" {
		t.Errorf("NetworkMainnet = %v, want mainnet", NetworkMainnet)
	}
	if NetworkPreprod != "preprod" {
		t.Errorf("NetworkPreprod = %v, want preprod", NetworkPreprod)
	}
	if NetworkPreview != "preview" {
		t.Errorf("NetworkPreview = %v, want preview", NetworkPreview)
	}
}

func TestTipStruct(t *testing.T) {
	tip := Tip{
		Slot:      1000000,
		Epoch:     100,
		Block:     500000,
		Hash:      "abc123",
		SyncState: "100.00",
	}

	if tip.Slot != 1000000 {
		t.Errorf("Tip.Slot = %v, want 1000000", tip.Slot)
	}
	if tip.Epoch != 100 {
		t.Errorf("Tip.Epoch = %v, want 100", tip.Epoch)
	}
	if tip.Block != 500000 {
		t.Errorf("Tip.Block = %v, want 500000", tip.Block)
	}
}

func TestUTxOStruct(t *testing.T) {
	utxo := UTxO{
		TxHash:  "tx123",
		TxIndex: 0,
		Address: "addr_test1abc",
		Value: Value{
			Lovelace: 10000000,
			Assets:   map[string]int64{"token1": 100},
		},
	}

	if utxo.TxHash != "tx123" {
		t.Errorf("UTxO.TxHash = %v, want tx123", utxo.TxHash)
	}
	if utxo.Value.Lovelace != 10000000 {
		t.Errorf("UTxO.Value.Lovelace = %v, want 10000000", utxo.Value.Lovelace)
	}
	if utxo.Value.Assets["token1"] != 100 {
		t.Errorf("UTxO.Value.Assets[token1] = %v, want 100", utxo.Value.Assets["token1"])
	}
}

func TestProtocolParametersStruct(t *testing.T) {
	params := ProtocolParameters{
		MinFeeA:             44,
		MinFeeB:             155381,
		MaxBlockSize:        90112,
		MaxTxSize:           16384,
		KeyDeposit:          2000000,
		PoolDeposit:         500000000,
		MaxEpoch:            18,
		NOpt:                500,
		PoolPledgeInfluence: "0.3",
		MinPoolCost:         340000000,
		SlotsPerKESPeriod:   129600,
		MaxKESEvolutions:    62,
	}

	if params.PoolDeposit != 500000000 {
		t.Errorf("PoolDeposit = %v, want 500000000", params.PoolDeposit)
	}
	if params.MaxEpoch != 18 {
		t.Errorf("MaxEpoch = %v, want 18", params.MaxEpoch)
	}
}

func TestPoolParamsStruct(t *testing.T) {
	params := PoolParams{
		PoolID:        "pool1abc",
		VRFKeyHash:    "vrf123",
		Pledge:        500000000000,
		Cost:          340000000,
		Margin:        0.03,
		RewardAccount: "stake1abc",
		Owners:        []string{"stake1abc"},
		Relays: []Relay{
			{Type: "dns", Hostname: "relay.example.com", Port: 6000},
		},
		MetadataURL:  "https://example.com/pool.json",
		MetadataHash: "hash123",
	}

	if params.Pledge != 500000000000 {
		t.Errorf("Pledge = %v, want 500000000000", params.Pledge)
	}
	if params.Margin != 0.03 {
		t.Errorf("Margin = %v, want 0.03", params.Margin)
	}
	if len(params.Relays) != 1 {
		t.Errorf("len(Relays) = %v, want 1", len(params.Relays))
	}
}

func TestKeyPairStruct(t *testing.T) {
	kp := KeyPair{
		SigningKey:      []byte("skey"),
		VerificationKey: []byte("vkey"),
	}

	if string(kp.SigningKey) != "skey" {
		t.Errorf("SigningKey = %v, want skey", string(kp.SigningKey))
	}
	if string(kp.VerificationKey) != "vkey" {
		t.Errorf("VerificationKey = %v, want vkey", string(kp.VerificationKey))
	}
}

func TestOperationalCertificateStruct(t *testing.T) {
	cert := OperationalCertificate{
		Certificate: []byte("cert-data"),
		Counter:     5,
		KESPeriod:   100,
		ExpiryEpoch: 200,
	}

	if cert.Counter != 5 {
		t.Errorf("Counter = %v, want 5", cert.Counter)
	}
	if cert.KESPeriod != 100 {
		t.Errorf("KESPeriod = %v, want 100", cert.KESPeriod)
	}
	if cert.ExpiryEpoch != 200 {
		t.Errorf("ExpiryEpoch = %v, want 200", cert.ExpiryEpoch)
	}
}

func TestTransactionSubmitResultStruct(t *testing.T) {
	result := TransactionSubmitResult{
		TxHash:  "tx-hash-abc",
		Success: true,
		Error:   "",
	}

	if result.TxHash != "tx-hash-abc" {
		t.Errorf("TxHash = %v, want tx-hash-abc", result.TxHash)
	}
	if !result.Success {
		t.Error("Success should be true")
	}

	failedResult := TransactionSubmitResult{
		TxHash:  "",
		Success: false,
		Error:   "insufficient funds",
	}

	if failedResult.Success {
		t.Error("Success should be false for failed result")
	}
	if failedResult.Error != "insufficient funds" {
		t.Errorf("Error = %v, want 'insufficient funds'", failedResult.Error)
	}
}

func TestClientConfigStruct(t *testing.T) {
	config := ClientConfig{
		Network:        NetworkPreprod,
		SocketPath:     "/var/run/cardano.sock",
		CLIPath:        "/usr/bin/cardano-cli",
		TimeoutSeconds: 60,
		NetworkMagic:   1,
	}

	if config.Network != NetworkPreprod {
		t.Errorf("Network = %v, want preprod", config.Network)
	}
	if config.SocketPath != "/var/run/cardano.sock" {
		t.Errorf("SocketPath = %v, want /var/run/cardano.sock", config.SocketPath)
	}
	if config.TimeoutSeconds != 60 {
		t.Errorf("TimeoutSeconds = %v, want 60", config.TimeoutSeconds)
	}
}

func TestTxBuildOptionsStruct(t *testing.T) {
	opts := TxBuildOptions{
		Inputs: []UTxO{
			{TxHash: "tx1", TxIndex: 0, Address: "addr1", Value: Value{Lovelace: 10000000}},
		},
		Outputs: []TxOutput{
			{Address: "addr2", Value: Value{Lovelace: 5000000}},
		},
		Certificates:  [][]byte{[]byte("cert1")},
		ChangeAddress: "addr1",
		TTLSlot:       1000000,
		Fee:           200000,
		Metadata:      map[string]interface{}{"key": "value"},
	}

	if len(opts.Inputs) != 1 {
		t.Errorf("len(Inputs) = %v, want 1", len(opts.Inputs))
	}
	if len(opts.Outputs) != 1 {
		t.Errorf("len(Outputs) = %v, want 1", len(opts.Outputs))
	}
	if opts.ChangeAddress != "addr1" {
		t.Errorf("ChangeAddress = %v, want addr1", opts.ChangeAddress)
	}
	if opts.Fee != 200000 {
		t.Errorf("Fee = %v, want 200000", opts.Fee)
	}
}

func TestTxOutputStruct(t *testing.T) {
	output := TxOutput{
		Address: "addr_test1abc",
		Value: Value{
			Lovelace: 5000000,
			Assets:   map[string]int64{"token1": 50},
		},
	}

	if output.Address != "addr_test1abc" {
		t.Errorf("Address = %v, want addr_test1abc", output.Address)
	}
	if output.Value.Lovelace != 5000000 {
		t.Errorf("Value.Lovelace = %v, want 5000000", output.Value.Lovelace)
	}
}

func TestRelayStruct(t *testing.T) {
	dnsRelay := Relay{
		Type:     "dns",
		Hostname: "relay.example.com",
		Port:     6000,
	}

	if dnsRelay.Type != "dns" {
		t.Errorf("Type = %v, want dns", dnsRelay.Type)
	}
	if dnsRelay.Hostname != "relay.example.com" {
		t.Errorf("Hostname = %v, want relay.example.com", dnsRelay.Hostname)
	}

	ipRelay := Relay{
		Type: "ip",
		IPv4: "192.168.1.1",
		IPv6: "2001:db8::1",
		Port: 6000,
	}

	if ipRelay.IPv4 != "192.168.1.1" {
		t.Errorf("IPv4 = %v, want 192.168.1.1", ipRelay.IPv4)
	}
	if ipRelay.IPv6 != "2001:db8::1" {
		t.Errorf("IPv6 = %v, want 2001:db8::1", ipRelay.IPv6)
	}
}
