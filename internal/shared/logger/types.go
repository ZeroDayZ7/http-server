package logger

type Env string

const (
	EnvProduction  Env = "production"
	EnvDevelopment Env = "development"
	EnvTest        Env = "test"
)
