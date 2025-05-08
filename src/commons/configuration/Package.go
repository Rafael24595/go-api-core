package configuration

type Package struct {
	Version string  `yaml:"version"`
	Project Project `yaml:"project"`
}

type Project struct {
    Name    string `yaml:"name"`
    Version string `yaml:"version"`
}
