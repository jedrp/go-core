package util

import "testing"

func TestGetConfig(t *testing.T) {
	tt := []struct {
		inputStr         string
		expectedKeyValue map[string]string
	}{
		{
			inputStr: "type=es;host=http://localhost;index-prefix=prefix;sniff=true;",
			expectedKeyValue: map[string]string{
				"type":         "es",
				"host":         "http://localhost",
				"index-prefix": "prefix",
				"sniff":        "true",
			},
		},
		{
			inputStr: "type=kafka;host=;enable=true",
			expectedKeyValue: map[string]string{
				"type": "kafka",
				"host": "",
			},
		},
	}
	for i, tc := range tt {
		c, err := GetConfig(tc.inputStr)
		if err != nil {
			t.Error(err)
		}
		for v := range tc.expectedKeyValue {
			if tc.expectedKeyValue[v] != c[v] {
				t.Errorf("tc #%d, expected %v but got %v", i, tc.expectedKeyValue[v], c[v])
			}
		}
	}

}
