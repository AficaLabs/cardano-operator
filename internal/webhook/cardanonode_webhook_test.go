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

package webhook

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cardanov1alpha1 "github.com/AficaLabs/cardano-operator/api/v1alpha1"
)

func newValidCardanoNode() *cardanov1alpha1.CardanoNode {
	return &cardanov1alpha1.CardanoNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-node",
			Namespace: "default",
		},
		Spec: cardanov1alpha1.CardanoNodeSpec{
			Type:    cardanov1alpha1.CardanoNodeTypeRelay,
			Network: cardanov1alpha1.NetworkPreprod,
			Storage: cardanov1alpha1.StorageConfig{
				Size: "200Gi",
			},
		},
	}
}

func TestCardanoNodeValidatorValidateCreate(t *testing.T) {
	ctx := context.Background()
	validator := &CardanoNodeValidator{}

	t.Run("valid relay node", func(t *testing.T) {
		cn := newValidCardanoNode()
		warnings, err := validator.ValidateCreate(ctx, cn)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateCreate() warnings = %v, want none", warnings)
		}
	})

	t.Run("valid block producer with stakePoolRef", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Type = cardanov1alpha1.CardanoNodeTypeBlockProducer
		cn.Spec.StakePoolRef = "my-pool"
		warnings, err := validator.ValidateCreate(ctx, cn)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateCreate() warnings = %v, want none", warnings)
		}
	})

	t.Run("block producer without stakePoolRef warning", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Type = cardanov1alpha1.CardanoNodeTypeBlockProducer
		cn.Spec.StakePoolRef = ""
		warnings, err := validator.ValidateCreate(ctx, cn)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
		if len(warnings) == 0 {
			t.Error("ValidateCreate() expected warning for block producer without stakePoolRef")
		}
	})

	t.Run("invalid object type", func(t *testing.T) {
		_, err := validator.ValidateCreate(ctx, &cardanov1alpha1.StakePool{})
		if err == nil {
			t.Error("ValidateCreate() expected error for wrong object type")
		}
	})

	t.Run("invalid node type", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Type = "invalid-type"
		_, err := validator.ValidateCreate(ctx, cn)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid node type")
		}
	})

	t.Run("invalid network", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Network = "invalid-network"
		_, err := validator.ValidateCreate(ctx, cn)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid network")
		}
	})

	t.Run("small storage warning", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Storage.Size = "50Gi"
		warnings, err := validator.ValidateCreate(ctx, cn)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
		if len(warnings) == 0 {
			t.Error("ValidateCreate() expected warning for small storage")
		}
	})

	t.Run("invalid storage size", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Storage.Size = "invalid"
		_, err := validator.ValidateCreate(ctx, cn)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid storage size")
		}
	})

	t.Run("valid p2p topology", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Topology = &cardanov1alpha1.TopologyConfig{
			Mode: cardanov1alpha1.TopologyModeP2P,
		}
		warnings, err := validator.ValidateCreate(ctx, cn)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateCreate() warnings = %v, want none", warnings)
		}
	})

	t.Run("valid static topology", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Topology = &cardanov1alpha1.TopologyConfig{
			Mode: cardanov1alpha1.TopologyModeStatic,
			StaticPeers: []cardanov1alpha1.Peer{
				{Address: "relay1.example.com", Port: 6000},
			},
		}
		warnings, err := validator.ValidateCreate(ctx, cn)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateCreate() warnings = %v, want none", warnings)
		}
	})

	t.Run("static topology without peers", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Topology = &cardanov1alpha1.TopologyConfig{
			Mode:        cardanov1alpha1.TopologyModeStatic,
			StaticPeers: []cardanov1alpha1.Peer{},
		}
		_, err := validator.ValidateCreate(ctx, cn)
		if err == nil {
			t.Error("ValidateCreate() expected error for static topology without peers")
		}
	})

	t.Run("static topology with invalid peer", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Topology = &cardanov1alpha1.TopologyConfig{
			Mode: cardanov1alpha1.TopologyModeStatic,
			StaticPeers: []cardanov1alpha1.Peer{
				{Address: "", Port: 6000},
			},
		}
		_, err := validator.ValidateCreate(ctx, cn)
		if err == nil {
			t.Error("ValidateCreate() expected error for peer without address")
		}
	})

	t.Run("static topology with invalid port", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Topology = &cardanov1alpha1.TopologyConfig{
			Mode: cardanov1alpha1.TopologyModeStatic,
			StaticPeers: []cardanov1alpha1.Peer{
				{Address: "relay.example.com", Port: 0},
			},
		}
		_, err := validator.ValidateCreate(ctx, cn)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid port")
		}
	})

	t.Run("invalid topology mode", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Topology = &cardanov1alpha1.TopologyConfig{
			Mode: "invalid",
		}
		_, err := validator.ValidateCreate(ctx, cn)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid topology mode")
		}
	})

	t.Run("offline-signing node type", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Type = cardanov1alpha1.CardanoNodeTypeOfflineSigning
		warnings, err := validator.ValidateCreate(ctx, cn)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateCreate() warnings = %v", warnings)
		}
	})
}

func TestCardanoNodeValidatorValidateUpdate(t *testing.T) {
	ctx := context.Background()
	validator := &CardanoNodeValidator{}

	t.Run("valid update", func(t *testing.T) {
		oldCN := newValidCardanoNode()
		newCN := newValidCardanoNode()
		newCN.Spec.Storage.Size = "300Gi"
		warnings, err := validator.ValidateUpdate(ctx, oldCN, newCN)
		if err != nil {
			t.Errorf("ValidateUpdate() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateUpdate() warnings = %v, want none", warnings)
		}
	})

	t.Run("invalid new object type", func(t *testing.T) {
		oldCN := newValidCardanoNode()
		_, err := validator.ValidateUpdate(ctx, oldCN, &cardanov1alpha1.StakePool{})
		if err == nil {
			t.Error("ValidateUpdate() expected error for wrong new object type")
		}
	})

	t.Run("invalid old object type", func(t *testing.T) {
		newCN := newValidCardanoNode()
		_, err := validator.ValidateUpdate(ctx, &cardanov1alpha1.StakePool{}, newCN)
		if err == nil {
			t.Error("ValidateUpdate() expected error for wrong old object type")
		}
	})

	t.Run("node type change rejected", func(t *testing.T) {
		oldCN := newValidCardanoNode()
		oldCN.Spec.Type = cardanov1alpha1.CardanoNodeTypeRelay
		newCN := newValidCardanoNode()
		newCN.Spec.Type = cardanov1alpha1.CardanoNodeTypeBlockProducer
		_, err := validator.ValidateUpdate(ctx, oldCN, newCN)
		if err == nil {
			t.Error("ValidateUpdate() expected error for node type change")
		}
	})

	t.Run("network change rejected", func(t *testing.T) {
		oldCN := newValidCardanoNode()
		oldCN.Spec.Network = cardanov1alpha1.NetworkPreprod
		newCN := newValidCardanoNode()
		newCN.Spec.Network = cardanov1alpha1.NetworkMainnet
		_, err := validator.ValidateUpdate(ctx, oldCN, newCN)
		if err == nil {
			t.Error("ValidateUpdate() expected error for network change")
		}
	})

	t.Run("storage size reduction warning", func(t *testing.T) {
		oldCN := newValidCardanoNode()
		oldCN.Spec.Storage.Size = "300Gi"
		newCN := newValidCardanoNode()
		newCN.Spec.Storage.Size = "200Gi"
		warnings, err := validator.ValidateUpdate(ctx, oldCN, newCN)
		if err != nil {
			t.Errorf("ValidateUpdate() error = %v", err)
		}
		if len(warnings) == 0 {
			t.Error("ValidateUpdate() expected warning for storage reduction")
		}
	})

	t.Run("small storage warning on update", func(t *testing.T) {
		oldCN := newValidCardanoNode()
		newCN := newValidCardanoNode()
		newCN.Spec.Storage.Size = "50Gi"
		warnings, err := validator.ValidateUpdate(ctx, oldCN, newCN)
		if err != nil {
			t.Errorf("ValidateUpdate() error = %v", err)
		}
		if len(warnings) == 0 {
			t.Error("ValidateUpdate() expected warning for small storage")
		}
	})

	t.Run("topology validation on update", func(t *testing.T) {
		oldCN := newValidCardanoNode()
		newCN := newValidCardanoNode()
		newCN.Spec.Topology = &cardanov1alpha1.TopologyConfig{
			Mode:        cardanov1alpha1.TopologyModeStatic,
			StaticPeers: []cardanov1alpha1.Peer{},
		}
		_, err := validator.ValidateUpdate(ctx, oldCN, newCN)
		if err == nil {
			t.Error("ValidateUpdate() expected error for invalid topology")
		}
	})
}

func TestCardanoNodeValidatorValidateDelete(t *testing.T) {
	ctx := context.Background()
	validator := &CardanoNodeValidator{}

	t.Run("delete relay node", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Type = cardanov1alpha1.CardanoNodeTypeRelay
		warnings, err := validator.ValidateDelete(ctx, cn)
		if err != nil {
			t.Errorf("ValidateDelete() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateDelete() warnings = %v, want none", warnings)
		}
	})

	t.Run("delete block producer warning", func(t *testing.T) {
		cn := newValidCardanoNode()
		cn.Spec.Type = cardanov1alpha1.CardanoNodeTypeBlockProducer
		warnings, err := validator.ValidateDelete(ctx, cn)
		if err != nil {
			t.Errorf("ValidateDelete() error = %v", err)
		}
		if len(warnings) == 0 {
			t.Error("ValidateDelete() expected warning for deleting block producer")
		}
	})

	t.Run("invalid object type", func(t *testing.T) {
		_, err := validator.ValidateDelete(ctx, &cardanov1alpha1.StakePool{})
		if err == nil {
			t.Error("ValidateDelete() expected error for wrong object type")
		}
	})
}

func TestValidateNodeType(t *testing.T) {
	tests := []struct {
		name     string
		nodeType cardanov1alpha1.CardanoNodeType
		wantErr  bool
	}{
		{"block-producer", cardanov1alpha1.CardanoNodeTypeBlockProducer, false},
		{"relay", cardanov1alpha1.CardanoNodeTypeRelay, false},
		{"offline-signing", cardanov1alpha1.CardanoNodeTypeOfflineSigning, false},
		{"invalid", "invalid", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNodeType(tt.nodeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNodeType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNodeNetwork(t *testing.T) {
	tests := []struct {
		name    string
		network cardanov1alpha1.Network
		wantErr bool
	}{
		{"mainnet", cardanov1alpha1.NetworkMainnet, false},
		{"preprod", cardanov1alpha1.NetworkPreprod, false},
		{"preview", cardanov1alpha1.NetworkPreview, false},
		{"invalid", "invalid", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNodeNetwork(tt.network)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNodeNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTopology(t *testing.T) {
	tests := []struct {
		name     string
		topology *cardanov1alpha1.TopologyConfig
		nodeType cardanov1alpha1.CardanoNodeType
		wantErr  bool
	}{
		{
			name: "p2p mode",
			topology: &cardanov1alpha1.TopologyConfig{
				Mode: cardanov1alpha1.TopologyModeP2P,
			},
			nodeType: cardanov1alpha1.CardanoNodeTypeRelay,
			wantErr:  false,
		},
		{
			name: "static mode with peers",
			topology: &cardanov1alpha1.TopologyConfig{
				Mode: cardanov1alpha1.TopologyModeStatic,
				StaticPeers: []cardanov1alpha1.Peer{
					{Address: "peer1.example.com", Port: 6000},
				},
			},
			nodeType: cardanov1alpha1.CardanoNodeTypeBlockProducer,
			wantErr:  false,
		},
		{
			name: "static mode without peers",
			topology: &cardanov1alpha1.TopologyConfig{
				Mode:        cardanov1alpha1.TopologyModeStatic,
				StaticPeers: []cardanov1alpha1.Peer{},
			},
			nodeType: cardanov1alpha1.CardanoNodeTypeRelay,
			wantErr:  true,
		},
		{
			name: "static mode with empty peer address",
			topology: &cardanov1alpha1.TopologyConfig{
				Mode: cardanov1alpha1.TopologyModeStatic,
				StaticPeers: []cardanov1alpha1.Peer{
					{Address: "", Port: 6000},
				},
			},
			nodeType: cardanov1alpha1.CardanoNodeTypeRelay,
			wantErr:  true,
		},
		{
			name: "static mode with invalid port low",
			topology: &cardanov1alpha1.TopologyConfig{
				Mode: cardanov1alpha1.TopologyModeStatic,
				StaticPeers: []cardanov1alpha1.Peer{
					{Address: "peer.example.com", Port: 0},
				},
			},
			nodeType: cardanov1alpha1.CardanoNodeTypeRelay,
			wantErr:  true,
		},
		{
			name: "static mode with invalid port high",
			topology: &cardanov1alpha1.TopologyConfig{
				Mode: cardanov1alpha1.TopologyModeStatic,
				StaticPeers: []cardanov1alpha1.Peer{
					{Address: "peer.example.com", Port: 70000},
				},
			},
			nodeType: cardanov1alpha1.CardanoNodeTypeRelay,
			wantErr:  true,
		},
		{
			name: "invalid topology mode",
			topology: &cardanov1alpha1.TopologyConfig{
				Mode: "invalid",
			},
			nodeType: cardanov1alpha1.CardanoNodeTypeRelay,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTopology(tt.topology, tt.nodeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTopology() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseStorageSizeGi(t *testing.T) {
	tests := []struct {
		name    string
		size    string
		want    int64
		wantErr bool
	}{
		{"200Gi", "200Gi", 200, false},
		{"200G", "200G", 200, false},
		{"1Ti", "1Ti", 1024, false},
		{"1T", "1T", 1000, false},
		{"500Mi", "500Mi", 0, false},
		{"500M", "500M", 0, false},
		{"1000Ki", "1000Ki", 0, false},
		{"1000K", "1000K", 0, false},
		{"1Pi", "1Pi", 1024 * 1024, false},
		{"1P", "1P", 1000 * 1000, false},
		{"1Ei", "1Ei", 1024 * 1024 * 1024, false},
		{"1E", "1E", 1000 * 1000 * 1000, false},
		{"empty", "", 0, true},
		{"invalid", "abc", 0, true},
		{"negative", "-100Gi", 0, true},
		{"whitespace", "  200Gi  ", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStorageSizeGi(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStorageSizeGi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseStorageSizeGi() = %v, want %v", got, tt.want)
			}
		})
	}
}
