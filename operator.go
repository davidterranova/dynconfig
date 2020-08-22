package dynconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// ConfigReader reads config
type ConfigReader interface {
	Read(config interface{}) error
}

// ConfigWriter writes config
type ConfigWriter interface {
	Write(config interface{}) error
}

// ConfigNotifier notifies if config changes
type ConfigNotifier interface {
	Register(operator *Operator)
	Watch(ctx context.Context) error
	ConfigReader
}

// Operator is responsible to coordinate readers, writers & notifiers
type Operator struct {
	config interface{}

	readers   []ConfigReader
	writers   []ConfigWriter
	notifiers []ConfigNotifier
}

// ConfigOption cinfigure the operator
type ConfigOption func(*Operator)

// WithConfigReader adds a reader to the operator
func WithConfigReader(cr ConfigReader) ConfigOption {
	return func(co *Operator) {
		co.readers = append(co.readers, cr)
	}
}

// WithConfigWriter adds a writer to the operator
func WithConfigWriter(cw ConfigWriter) ConfigOption {
	return func(co *Operator) {
		co.writers = append(co.writers, cw)
	}
}

// WithConfigNotifier adds a notifier to the operator
func WithConfigNotifier(cn ConfigNotifier) ConfigOption {
	return func(co *Operator) {
		co.notifiers = append(co.notifiers, cn)
		cn.Register(co)
	}
}

// NewOperator creates an operator
func NewOperator(config interface{}, opts ...ConfigOption) *Operator {
	co := &Operator{
		config:    config,
		readers:   make([]ConfigReader, 0),
		writers:   make([]ConfigWriter, 0),
		notifiers: make([]ConfigNotifier, 0),
	}

	for _, opt := range opts {
		opt(co)
	}

	return co
}

// ConfigChanged is being called by notifiers
func (o Operator) ConfigChanged(cr ConfigReader) {
	if err := cr.Read(o.config); err != nil {
		log.Printf("failed to read config: %s", err)
	}
	if err := o.write(o.config); err != nil {
		log.Printf("failed to write config: %s", err)
	}
	data, _ := json.Marshal(o.config)
	log.Printf("config changed: %s", string(data))
}

func (o Operator) read(config interface{}) error {
	for _, reader := range o.readers {
		if err := reader.Read(config); err != nil {
			return fmt.Errorf("config read failed: %w", err)
		}
	}
	return nil
}

func (o Operator) write(config interface{}) error {
	for _, writer := range o.writers {
		if err := writer.Write(config); err != nil {
			return fmt.Errorf("config write failed: %w", err)
		}
	}
	return nil
}

func (o Operator) startWatchers(ctx context.Context) error {
	for _, notifier := range o.notifiers {
		if err := notifier.Watch(ctx); err != nil {
			return fmt.Errorf("failed to start notifier: %w", err)
		}
	}
	return nil
}

// Process
// - read config from readers
// - write config to writers
// - start watchers
func (o Operator) Process(ctx context.Context) error {
	if err := o.read(o.config); err != nil {
		return err
	}
	if err := o.write(o.config); err != nil {
		return err
	}
	if err := o.startWatchers(ctx); err != nil {
		return err
	}
	return nil
}
