package config

import "time"

// Config is the root configuration object loaded from the active env JSON file
// (overlaid with environment variables via Viper).
type Config struct {
	App      App      `mapstructure:"app"`
	Server   Server   `mapstructure:"server"`
	Store    Store    `mapstructure:"store"`
	Cache    Cache    `mapstructure:"cache"`
	Token    Token    `mapstructure:"token"`
	WhatsApp WhatsApp `mapstructure:"whatsapp"`
	Auth     Auth     `mapstructure:"auth"`
	MsgQueue MsgQueue `mapstructure:"msgQueue"`
}

// App holds high-level service metadata.
type App struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

// Server holds HTTP server tuning.
type Server struct {
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdownTimeout"`
}

// Store holds the primary SQL (PostgreSQL) connection settings. Same shape as
// auth-service, pointed at a separate database (expense_service).
type Store struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"sslMode"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	AutoMigrate     bool          `mapstructure:"autoMigrate"`
}

// Cache holds the Redis connection settings. This is the SAME Redis instance the
// auth-service uses. Keys are namespaced by an "expense:" prefix to avoid
// colliding with auth-service's keys, so both services share DB index 0 —
// required by managed Redis (Upstash) which only supports the 0th database.
type Cache struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	// TLS enables an encrypted connection — required by managed Redis providers
	// (Upstash, Redis Cloud). Leave false for a local plaintext Redis.
	TLS bool `mapstructure:"tls"`
	// SummaryTTL is how long the dashboard analytics summaries are cached.
	SummaryTTL time.Duration `mapstructure:"summaryTTL"`
}

// Token holds the JWT verification settings. The expense-service only ever
// VERIFIES access tokens issued by auth-service, so it loads the RSA public key
// (never the private key) and matches the issuer.
type Token struct {
	Issuer        string `mapstructure:"issuer"`
	PublicKeyPath string `mapstructure:"publicKeyPath"`
	PublicKeyPEM  string `mapstructure:"publicKeyPEM"`
}

// WhatsApp holds Meta Cloud API settings for the inbound webhook (verification +
// signature) and the outbound message client. Secrets come from env overrides.
type WhatsApp struct {
	// VerifyToken is echoed back during Meta's GET webhook handshake.
	VerifyToken string `mapstructure:"verifyToken"`
	// AppSecret signs inbound payloads (X-Hub-Signature-256, HMAC-SHA256).
	AppSecret string `mapstructure:"appSecret"`
	// AccessToken authorizes outbound sends to the Graph API.
	AccessToken string `mapstructure:"accessToken"`
	// PhoneNumberID is the sender's Cloud API phone-number id.
	PhoneNumberID string `mapstructure:"phoneNumberID"`
	// GraphBaseURL + GraphVersion build the send endpoint; defaulted in config.
	GraphBaseURL string `mapstructure:"graphBaseURL"`
	GraphVersion string `mapstructure:"graphVersion"`
	// SignupURL is shown to unregistered senders.
	SignupURL string `mapstructure:"signupURL"`
}

// Auth holds the auth-service base URL and the shared internal key used to call
// its /v1/internal routes (phone → user lookup).
type Auth struct {
	URL         string `mapstructure:"url"`
	InternalKey string `mapstructure:"internalKey"`
}

// MsgQueue holds Kafka settings for the async WhatsApp pipeline. When Enabled is
// false the ingress/worker fall back to a logging stub so local dev needs no
// broker.
type MsgQueue struct {
	Brokers       []string `mapstructure:"brokers"`
	InboundTopic  string   `mapstructure:"inboundTopic"`
	DLQTopic      string   `mapstructure:"dlqTopic"`
	ConsumerGroup string   `mapstructure:"consumerGroup"`
	Enabled       bool     `mapstructure:"enabled"`
}
