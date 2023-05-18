package configuration

type TransportConfiguration struct {
	Port           string   `json:"port"`
	ReadTimeOutMs  string   `json:"readTimeOutMs,omitempty"`
	WriteTimeOutMs string   `json:"writeTimeOutMs,omitempty"`
	ContextHeaders []string `json:"contextHeaders,omitempty"`

	DiscoveryServiceProvider *string     `json:"discoveryServiceProvider"`
	DspTemplate              *string     `json:"dspTemplate,omitempty"`
	AssetRoots               interface{} `json:"assetRoots,omitempty"`
}
