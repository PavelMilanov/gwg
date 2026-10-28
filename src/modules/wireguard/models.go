package wireguard

type WgServerConfig struct {
	ServerPrivateKey string
	ServerPublicKey  string
	LocalAddress     string
	PublicAddress    string
	ListenPort       int
	Eth              string
	Alias            string
	DnsResolv        string
	Users            []UserConfig // for peer
}

type UserConfig struct {
	ClientPrivateKey   string
	ClientPublicKey    string
	ClientLocalAddress string
	ServerPublicKey    string
	ServerIp           string
	ServerPort         int
	DnsResolv          string
	Name               string
	Status             string
}
