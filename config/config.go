package config

var Port = 1324

type AwsConfig struct {
	AwsId        string
	AwsSecretKey string
	AwsRegion    string
}

type Config struct {
	Status    bool
	AwsConfig AwsConfig
}

var Conf Config

//func Load(location string) (conf Config, err error) {
//	var reader io.Reader
//	// check for http prefix
//
//	log.Infof("loading local config (%v)", location)
//
//	// check the conf file exists
//	if _, err := os.Stat(location); os.IsNotExist(err) {
//		return conf, fmt.Errorf("config file at location (%v) not found!", location)
//	}
//	// open the config file
//	reader, err = os.Open(location)
//
//	if err != nil {
//		return conf, fmt.Errorf("error opening local config file (%v): %v ", location, err)
//	}
//	_, err = toml.DecodeReader(reader, &conf)
//
//	return conf, err
//}
