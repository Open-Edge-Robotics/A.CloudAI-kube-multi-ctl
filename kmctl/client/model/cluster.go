package model

type Clusters struct {
	Cluster []Cluster `mapstructure:"server"`
}

type Cluster struct {
	Name string `mapstructure:"name"`
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}
