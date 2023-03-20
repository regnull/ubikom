package cfg

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ConfigEntry interface {
	SetDefault()
	BindEnv()
	BindFlag()
}
type ConfigBase[T any] struct {
	Name         string
	Value        T
	DefaultValue T
	FlagHelp     string
	EnvVar       string
}

func (cb *ConfigBase[T]) SetDefault() {
	viper.SetDefault(cb.Name, cb.DefaultValue)
}

func (cb *ConfigBase[T]) BindEnv() {
	if cb.EnvVar == "" {
		return
	}
	viper.BindEnv(cb.Name, cb.EnvVar)
}

type StringConfig struct {
	ConfigBase[string]
}

func NewStringConfig(name string, defaultValue string, flagHelp string, envVar string) ConfigEntry {
	return &StringConfig{
		ConfigBase: ConfigBase[string]{
			Name:         name,
			DefaultValue: defaultValue,
			FlagHelp:     flagHelp,
			EnvVar:       envVar,
		},
	}
}

func (sc *StringConfig) BindFlag() {
	flag.StringVar(&sc.Value, sc.Name, "", sc.FlagHelp)
}

type IntConfig struct {
	ConfigBase[int]
}

func NewIntConfig(name string, defaultValue int, flagHelp string, envVar string) ConfigEntry {
	return &IntConfig{
		ConfigBase: ConfigBase[int]{
			Name:         name,
			DefaultValue: defaultValue,
			FlagHelp:     flagHelp,
			EnvVar:       envVar,
		},
	}
}

func (ic *IntConfig) BindFlag() {
	flag.IntVar(&ic.Value, ic.Name, 0, ic.FlagHelp)
}

type BoolConfig struct {
	ConfigBase[bool]
}

func NewBoolConfig(name string, defaultValue bool, flagHelp string, envVar string) ConfigEntry {
	return &BoolConfig{
		ConfigBase: ConfigBase[bool]{
			Name:         name,
			DefaultValue: defaultValue,
			FlagHelp:     flagHelp,
			EnvVar:       envVar,
		},
	}
}

func (bc *BoolConfig) BindFlag() {
	flag.BoolVar(&bc.Value, bc.Name, false, bc.FlagHelp)
}

func InitConfig(entries []ConfigEntry) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var configFile string
	flag.StringVar(&configFile, "config", "", "config file location")

	for _, e := range entries {
		e.SetDefault()
		e.BindEnv()
		e.BindFlag()
	}

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	if configFile != "" {
		viper.SetConfigFile(configFile)
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	return nil
}
