package elasticsearch

type Config struct {
	Addrs    []string `config:"addrs" json:"addrs"`
	Username string   `config:"username" json:"username"`
	Password string   `config:"password" json:"password"`
}
