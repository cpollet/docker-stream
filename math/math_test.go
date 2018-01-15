package math

import "testing"

func TestMax(t *testing.T) {
	var testCases = []struct {
		X        int
		Y        int
		Expected int
	}{
		{
			X:        0,
			Y:        1,
			Expected: 1,
		},
		{
			X:        -1,
			Y:        1,
			Expected: 1,
		},
		{
			X:        -2,
			Y:        -1,
			Expected: -1,
		},
	}

	for _, testCase := range testCases {
		actual := Max(testCase.X, testCase.Y)
		if actual != testCase.Expected {
			t.Errorf("Expected Min(%d, %d) to give %d, got %d", testCase.X, testCase.Y, testCase.Expected, actual)
		}
	}
}
