package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Fabric   Fabric   `json:"fabric"`
	Database Database `json:"database"`
}
type PeerNode struct {
	Endpoint string `yaml:"endpoint"`
	HostName string `yaml:"host_name"`
}

type Fabric struct {
	MspID           string     `yaml:"msp_id"`
	ChannelName     string     `yaml:"channel_name"`
	ChaincodeName   string     `yaml:"chaincode_name"`
	CertPath        string     `yaml:"cert_path"`
	KeyDir          string     `yaml:"key_dir"`
	TLSRootCertPath string     `yaml:"tls_root_cert_path"`
	Peers           []PeerNode `yaml:"peers"`
}

type Database struct {
	MongoURI   string `yaml:"mongo_uri"`
	DBName     string `yaml:"db_name"`
	Collection string `yaml:"collection"`
}

func Load(cfgFile string) (*Config, error) {
	cfgData, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config:%v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(cfgData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to resolve config data:%v", err)
	}
	return &cfg, nil
}
