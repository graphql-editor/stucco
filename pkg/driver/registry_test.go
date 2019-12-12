package driver_test

import (
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/drivertest"
	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	data := []struct {
		title           string
		registerConfig  driver.Config
		registerDriver  driver.Driver
		getDriverConfig driver.Config
		getDriverDriver driver.Driver
	}{
		{
			title: "ReturnsRegisteredDriver",
			registerConfig: driver.Config{
				Provider: "provider",
				Runtime:  "runtime",
			},
			registerDriver: new(drivertest.MockDriver),
			getDriverConfig: driver.Config{
				Provider: "provider",
				Runtime:  "runtime",
			},
			getDriverDriver: new(drivertest.MockDriver),
		},
		{
			title: "ReturnsNilOnNotFound",
			registerConfig: driver.Config{
				Provider: "provider",
				Runtime:  "runtime",
			},
			registerDriver: new(drivertest.MockDriver),
			getDriverConfig: driver.Config{
				Provider: "provider2",
				Runtime:  "runtime",
			},
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			driver.Register(tt.registerConfig, tt.registerDriver)
			assert.Equal(t, tt.getDriverDriver, driver.GetDriver(tt.getDriverConfig))
		})
	}
}
