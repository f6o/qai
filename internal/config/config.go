package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Pomodoro PomodoroConfig `mapstructure:"pomodoro"`
	Data     DataConfig     `mapstructure:"data"`
}

type PomodoroConfig struct {
	WorkMinutes  int  `mapstructure:"work_minutes"`
	BreakMinutes int  `mapstructure:"break_minutes"`
	Notify       bool `mapstructure:"notify"`
}

type DataConfig struct {
	Todofile    string `mapstructure:"todofile"`
	Logfile     string `mapstructure:"logfile"`
	MarkdownDir string `mapstructure:"markdowndir"`
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	qaiDir := filepath.Join(home, ".config", "qai")

	return &Config{
		Pomodoro: PomodoroConfig{
			WorkMinutes:  25,
			BreakMinutes: 5,
			Notify:       true,
		},
		Data: DataConfig{
			Todofile:    filepath.Join(qaiDir, "tasks.yaml"),
			Logfile:     filepath.Join(qaiDir, "logs.jsonl"),
			MarkdownDir: filepath.Join(qaiDir, "markdown"),
		},
	}
}

func Load() (*Config, error) {
	v := viper.New()

	home, _ := os.UserHomeDir()
	qaiDir := filepath.Join(home, ".config", "qai")

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(qaiDir)
	v.AddConfigPath(".")

	cfg := Default()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return cfg, nil
		}
		return nil, err
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Save() error {
	home, _ := os.UserHomeDir()
	qaiDir := filepath.Join(home, ".config", "qai")

	if err := os.MkdirAll(qaiDir, 0755); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType("toml")
	v.Set("pomodoro.work_minutes", c.Pomodoro.WorkMinutes)
	v.Set("pomodoro.break_minutes", c.Pomodoro.BreakMinutes)
	v.Set("pomodoro.notify", c.Pomodoro.Notify)
	v.Set("data.todofile", c.Data.Todofile)
	v.Set("data.logfile", c.Data.Logfile)
	v.Set("data.markdowndir", c.Data.MarkdownDir)

	return v.SafeWriteConfigAs(filepath.Join(qaiDir, "config.toml"))
}

func (c *Config) EnsureDirectories() error {
	dirs := []string{
		filepath.Dir(c.Data.Todofile),
		filepath.Dir(c.Data.Logfile),
		c.Data.MarkdownDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
