package database

type DbConfig struct {
	User     string `required:"true" split_words:"true"`
	Password string `required:"true" split_words:"true"`
	Host     string `required:"true" split_words:"true"`
	Port     int    `required:"true" split_words:"true"`
	Name     string `required:"true" split_words:"true"`
}
