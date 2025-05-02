package measurements

import "github.com/DataDog/datadog-go/v5/statsd"

func NewClient() Measurer {
	stats, err := statsd.
		New("",
			statsd.WithNamespace("backend"),
			statsd.WithTags([]string{"application:proxy"}),
		)
	if err != nil {
		println("DogStatsD not found, ignoring...", err.Error())
	}

	return NewMeasurementClient(stats)
}
