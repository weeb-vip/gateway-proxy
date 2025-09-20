package measurements

type Measurer interface {
	// MeasureExecutionTime when deferred will measure and send execution time of that method body
	MeasureExecutionTime(label string, tags []string) func()
	MarkEvent(label string, tags []string)
	MarkEventWithCount(label string, count int, tags []string)
}

type measurement struct {
}

func NewMeasurementClient() Measurer {
	return measurement{}
}

func (m measurement) MeasureExecutionTime(label string, tags []string) func() {
	return func() {} // no-op
}

func (m measurement) MarkEvent(label string, tags []string) {
	// no-op
}

func (m measurement) MarkEventWithCount(label string, count int, tags []string) {
	// no-op
}
