package fabric

type PeerNode struct {
	Endpoint string
	HostName string
}

type FabricConfig struct {
	MspID           string     `yaml:"msp_id"`
	ChannelName     string     `yaml:"channel_name"`
	ChaincodeName   string     `yaml:"chaincode_name"`
	CertPath        string     `yaml:"cert_path"`
	KeyDir          string     `yaml:"key_dir"`
	TlSRootCertPath string     `yaml:"tls_root_cert_path"`
	Peers           []PeerNode `yaml:"peers"`
}

type ConfigWrapper struct {
	FConfig FabricConfig `yaml:"fabric"`
}
