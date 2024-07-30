package utils

type Config struct {
  App *App `yaml:"app"`
}

type App struct {
  Port     int    `yaml:"port"`
  FilePath string `yaml:"filePath"`
  Password string `yaml:"password"`
}
