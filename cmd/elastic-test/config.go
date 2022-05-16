package main

type Config struct {
	Addrs     []string `config:"addrs" json:"addrs"`
	IndexName string   `config:"index_name" json:"index_name"`
	IndexType string   `config:"index_type" json:"index_type"`
	Username  string   `config:"username" json:"username"`
	Password  string   `config:"password" json:"password"`
}
