package configuration

var instance *Configuration

type Configuration struct {
	kargs map[string]string
}

func Initialize(kargs map[string]string) Configuration {
	if instance != nil {
		panic("")
	}
	instance = &Configuration{
		kargs: kargs,
	}

	return *instance
}

func Instance() Configuration {
	if instance == nil {
		panic("")
	}
	return *instance
}