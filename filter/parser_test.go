package filter

import "testing"

func TestMachineFilter(t *testing.T) {
	fixture := []string{
		"cpu < 10",
		"cpu = 10",
		"mem >= 10",
		"cpu != 10",
		"cpu < 10 & mem > 10",
		"cpu > 10 & ( mem < 10 | mem > 100)",
		"cpu < 10 | cpu > 100",
		"name = 'foo' & cpu < 10",
		"name != 'foo' & cpu < 10",
		"name ~= 'foo' & cpu < 10",
	}

	for _, expr := range fixture {
		_, err := parseMachineFilter(expr)
		if err != nil {
			t.Error(expr, err)
		}
	}
}
