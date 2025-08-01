package config

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		DBName   string `yaml:"db_name"`
	} `yaml:"database"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	UserService struct {
		URL string `yaml:"url"`
	} `yaml:"user_service"`

	AuthService struct {
		URL string `yaml:"url"`
	} `yaml:"auth_service"`
}
