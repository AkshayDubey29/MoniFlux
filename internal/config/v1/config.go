package v1

// Config holds the configuration for the application
type Config struct {
    APIPort    string `mapstructure:"api_port"`
    LogLevel   string `mapstructure:"log_level"`
    MongoURI   string `mapstructure:"mongo_uri"`
    MongoDB    string `mapstructure:"mongo_db"`
    // Add other configuration fields as necessary
}
