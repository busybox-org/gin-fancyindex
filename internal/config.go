package internal

type Config struct {
	Addr         string `envconfig:"ADDR" default:""`
	Port         string `envconfig:"PORT" default:"8080"`
	RelativePath string `envconfig:"RELATIVE_PATH" default:"/"`
	Root         string `envconfig:"ROOT" default:"/share"`
	Auth         bool   `envconfig:"AUTH" default:"false"`
	User         string `envconfig:"USER" default:"admin"`
	Pass         string `envconfig:"PASS" default:"admin"`
}
