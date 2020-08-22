package dynconfig_test

import (
	"context"
	"log"

	"github.com/davidterranova/dynconfig"
	"github.com/davidterranova/dynconfig/adaptors"
)

// Example_dynconfig
// with this environment variable set
//
//   os.Setenv("MYAPP_HOST", "127.0.0.1")
func Example_dynconfig() {
	type Config struct {
		Host string `envconfig:"HOST" default:"0.0.0.0"`
		Port int    `envconfig:"PORT" default:"8080"`
	}

	var config = Config{}

	// read from environment variables
	envAdaptor := adaptors.NewEnvConfigAdaptor("MY_APP")

	// read / write config from / to yaml file
	yamlAdaptor := adaptors.NewYAMLFileAdatptor("/etc/my_app/config.yaml")

	// will check for changes and update config accordingly
	fileWatcher := adaptors.NewFileWatcherAdaptor(
		"/etc/my_app/config.yaml",
		yamlAdaptor,
	)

	// create the operator
	operator := dynconfig.NewOperator(
		&config,
		dynconfig.WithConfigReader(envAdaptor),
		dynconfig.WithConfigReader(yamlAdaptor),
		dynconfig.WithConfigWriter(yamlAdaptor),
		dynconfig.WithConfigNotifier(fileWatcher),
	)

	// get configuration & starts watchers
	if err := operator.Process(context.Background()); err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", config)
	// Output : filled Config if the file exist
}
