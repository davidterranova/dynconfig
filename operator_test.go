package dynconfig_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/davidterranova/dynconfig"
	"github.com/davidterranova/dynconfig/adaptors"
	"github.com/stretchr/testify/assert"
)

type Config struct {
	Host string `envconfig:"HOST" default:"127.0.0.1"`
	Port int    `envconfig:"PORT" default:"80"`
}

func TestOperator(t *testing.T) {
	var testCases = []struct {
		name           string
		setup          func()
		tearDown       func()
		inputConfig    Config
		inputOptions   []ConfigOption
		expectedError  error
		expectedConfig Config
	}{
		{
			name:          "envconfigNoVariables",
			inputConfig:   Config{},
			inputOptions:  []ConfigOption{WithConfigReader(adaptors.NewEnvConfigAdaptor("HOTRELOAD"))},
			expectedError: nil,
			expectedConfig: Config{
				Host: "127.0.0.1",
				Port: 80,
			},
		},
		{
			name: "envconfigWithNonZeroValue",
			inputConfig: Config{
				Port: 5050,
			},
			inputOptions:  []ConfigOption{WithConfigReader(adaptors.NewEnvConfigAdaptor("HOTRELOAD"))},
			expectedError: nil,
			expectedConfig: Config{
				Host: "127.0.0.1",
				Port: 5050,
			},
		},
		{
			name: "envconfigWithEnv",
			setup: func() {
				os.Clearenv()
				os.Setenv("HOTRELOAD_PORT", "5050")
			},
			tearDown: func() {
				os.Clearenv()
			},
			inputConfig:   Config{},
			inputOptions:  []ConfigOption{WithConfigReader(adaptors.NewEnvConfigAdaptor("HOTRELOAD"))},
			expectedError: nil,
			expectedConfig: Config{
				Host: "127.0.0.1",
				Port: 5050,
			},
		},
		{
			name:           "noFileYAMLFileReader",
			inputConfig:    Config{},
			inputOptions:   []ConfigOption{WithConfigReader(adaptors.NewYAMLFileAdatptor("tmp/config.yaml"))},
			expectedError:  nil,
			expectedConfig: Config{},
		},
		{
			name: "emptyFileYAMLFileReader",
			setup: func() {
				file := "tmp/config.yaml"
				os.MkdirAll(file, 0700)
				ioutil.WriteFile(file, []byte{}, 0700)
			},
			tearDown: func() {
				file := "tmp/config.yaml"
				os.RemoveAll(file)
			},
			inputConfig:    Config{},
			inputOptions:   []ConfigOption{WithConfigReader(adaptors.NewYAMLFileAdatptor("tmp/config.yaml"))},
			expectedError:  nil,
			expectedConfig: Config{},
		},
		{
			name: "existingFileYAMLFileReader",
			setup: func() {
				file := "tmp/config.yaml"
				ioutil.WriteFile(file, []byte(`host: 0.0.0.0
`), 0700)
			},
			tearDown: func() {
				file := "tmp/config.yaml"
				os.RemoveAll(file)
			},
			inputConfig:   Config{},
			inputOptions:  []ConfigOption{WithConfigReader(adaptors.NewYAMLFileAdatptor("tmp/config.yaml"))},
			expectedError: nil,
			expectedConfig: Config{
				Host: "0.0.0.0",
			},
		},
		{
			name: "noFileYAMLFileWriter",
			tearDown: func() {
				os.Remove("tmp/config.yaml")
			},
			inputConfig:    Config{},
			inputOptions:   []ConfigOption{WithConfigWriter(adaptors.NewYAMLFileAdatptor("tmp/config.yaml"))},
			expectedError:  nil,
			expectedConfig: Config{},
		},
		{
			name: "noFileYAMLFileWriter",
			setup: func() {
				os.Setenv("HOTRELOAD_PORT", "5050")
				ioutil.WriteFile("tmp/config.yaml", []byte(`host: 0.0.0.0
`), 0700)
			},
			tearDown: func() {
				os.Remove("tmp/config.yaml")
			},
			inputConfig: Config{},
			inputOptions: []ConfigOption{
				WithConfigReader(adaptors.NewEnvConfigAdaptor("HOTRELOAD")),
				WithConfigReader(adaptors.NewYAMLFileAdatptor("tmp/config.yaml")),
				WithConfigWriter(adaptors.NewYAMLFileAdatptor("tmp/config.yaml")),
			},
			expectedError: nil,
			expectedConfig: Config{
				Host: "0.0.0.0",
				Port: 5050,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			operator := NewOperator(&testCase.inputConfig, testCase.inputOptions...)
			err := operator.Process(context.Background())
			if testCase.expectedError != nil {
				assert.NotNil(t, err)
				return
			}
			assert.Equal(t, testCase.expectedConfig, testCase.inputConfig)
			if testCase.tearDown != nil {
				testCase.tearDown()
			}
		})
	}
}

type notifier struct {
	o *Operator
}

func (n *notifier) Register(o *Operator)            { n.o = o }
func (n *notifier) Watch(ctx context.Context) error { return nil }
func (n *notifier) Read(config interface{}) error {
	c := config.(*Config)
	c.Host = "changed"
	return nil
}

func TestNotifier(t *testing.T) {
	var cfg = Config{}
	n := &notifier{}
	operator := NewOperator(
		&cfg,
		WithConfigNotifier(n),
	)
	operator.ConfigChanged(n)

	assert.Equal(t, "changed", cfg.Host)
}
