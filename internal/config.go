package internal

type Config struct {
	Addr         string `envconfig:"ADDR" default:""`
	Port         string `envconfig:"PORT" default:"8080"`
	RelativePath string `envconfig:"RELATIVE_PATH" default:"/"`
	Root         string `envconfig:"ROOT" default:"/share"`
	Auth         bool   `envconfig:"AUTH" default:"false"`
	AuthUser     string `envconfig:"AUTH_USER" default:"admin"`
	AuthPass     string `envconfig:"AUTH_PASS" default:"admin"`
}
