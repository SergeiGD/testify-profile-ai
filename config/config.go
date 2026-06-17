package config

import "time"

type Config struct {
	App struct {
		Name    string `yaml:"name" env:"APP_NAME"`
		Version string `yaml:"version" env:"APP_VERSION"`
		Debug   bool   `yaml:"Debug" env:"DEBUG"`
		BaseURL string `yaml:"base_url" env:"APP_BASE_URL"`
	}
	Database struct {
		Host       string        `env:"DATABASE_HOST"`
		User       string        `env:"DATABASE_USER"`
		Password   string        `env:"DATABASE_PASSWORD"`
		Port       string        `env:"DATABASE_PORT"`
		Db         string        `env:"DATABASE_NAME"`
		Timeout    time.Duration `env:"DATABASE_TIMEOUT"`
		ConnDelay  time.Duration `env:"DATABASE_DELAY"`
		MaxAttemps int           `env:"DATABASE_MAXATTEMPS"`
	}
	SMTP struct {
		Host          string `env:"SMTP_HOST"`
		Port          int    `env:"SMTP_PORT"`
		Username      string `env:"SMTP_USERNAME"`
		Password      string `env:"SMTP_PASSWORD"`
		From          string `env:"SMTP_FROM"`
		UseMockSender bool   `env:"USE_MOCK_EMAIL_SENDER" env-default:"false"`
	}
	Registration struct {
		TokenTTL       time.Duration `env:"REGISTRATION_TOKEN_TTL"`
		ResendCooldown time.Duration `env:"REGISTRATION_RESEND_COOLDOWN"`
	}
	JWT struct {
		PrivateKey string        `env:"JWT_PRIVATE_KEY"`
		PublicKey  string        `env:"JWT_PUBLIC_KEY"`
		AccessTTL  time.Duration `env:"JWT_ACCESS_TTL"`
		RefreshTTL time.Duration `env:"JWT_REFRESH_TTL"`
	}
}
