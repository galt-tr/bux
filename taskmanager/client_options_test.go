package taskmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWithNewRelic will test the method WithNewRelic()
func TestWithNewRelic(t *testing.T) {

	t.Run("check type", func(t *testing.T) {
		opt := WithNewRelic()
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying", func(t *testing.T) {
		options := &clientOptions{}
		opt := WithNewRelic()
		opt(options)
		assert.Equal(t, true, options.newRelicEnabled)
	})
}

// TestWithDebugging will test the method WithDebugging()
func TestWithDebugging(t *testing.T) {

	t.Run("check type", func(t *testing.T) {
		opt := WithDebugging()
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying", func(t *testing.T) {
		options := &clientOptions{}
		opt := WithDebugging()
		opt(options)
		assert.Equal(t, true, options.debug)
	})
}

// TestWithTaskQ will test the method WithTaskQ()
func TestWithTaskQ(t *testing.T) {
	t.Run("check type", func(t *testing.T) {
		opt := WithTaskQ(nil, FactoryEmpty)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil config", func(t *testing.T) {
		options := &clientOptions{
			taskq: &taskqOptions{
				config:      nil,
				factory:     nil,
				factoryType: "",
				queue:       nil,
				tasks:       nil,
			},
		}
		opt := WithTaskQ(nil, FactoryEmpty)
		opt(options)
		assert.Equal(t, Factory(""), options.taskq.factoryType)
		assert.Nil(t, options.taskq.config)
	})

	t.Run("test applying valid config", func(t *testing.T) {
		options := &clientOptions{
			taskq: &taskqOptions{},
		}
		opt := WithTaskQ(DefaultTaskQConfig(testQueueName), FactoryMemory)
		opt(options)
		assert.Equal(t, FactoryMemory, options.taskq.factoryType)
		assert.NotNil(t, options.taskq.config)
		assert.Equal(t, TaskQ, options.engine)
	})
}

// TestWithLogger will test the method WithLogger()
func TestWithLogger(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithLogger(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{}
		opt := WithLogger(nil)
		opt(options)
		assert.Nil(t, options.logger)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{}
		customClient := newLogger()
		opt := WithLogger(customClient)
		opt(options)
		assert.Equal(t, customClient, options.logger)
	})
}
