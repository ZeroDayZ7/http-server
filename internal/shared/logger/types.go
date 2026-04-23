package logger

type Env string

const (
	EnvProduction  Env = "production"
	EnvStaging     Env = "staging"
	EnvDevelopment Env = "development"
	EnvTest        Env = "test"
)
