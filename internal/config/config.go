package config

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
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

	OrderService struct {
		URL string `yaml:"url"`
	} `yaml:"order_service"`

	BillingService struct {
		URL string `yaml:"url"`
	} `yaml:"billing_service"`

	InventoryService struct {
		URL string `yaml:"url"`
	} `yaml:"inventory_service"`

	DeliveryService struct {
		URL string `yaml:"url"`
	} `yaml:"delivery_service"`

	RedpandaBroker struct {
		Addresses               []string `yaml:"addresses"`
		NewUsersTopic           string   `yaml:"new_users_topic"`
		NewOrdersTopic          string   `yaml:"new_orders_topic"`
		OrderIsPaidTopic        string   `yaml:"order_is_paid_topic"`
		OrderPaymentFailedTopic string   `yaml:"order_payment_failed_topic"`
		ConsumerGroup           string   `yaml:"consumer_group"`
	} `yaml:"redpanda_broker"`
}
