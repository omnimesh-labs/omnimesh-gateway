package helpers

import (
	"fmt"
	"reflect"
	"testing"
)

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		msg := fmt.Sprintf("Expected: %v, Actual: %v", expected, actual)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
}

// AssertNotEqual asserts that two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		msg := fmt.Sprintf("Expected not equal to: %v", expected)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
}

// AssertNil asserts that a value is nil
func AssertNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if value != nil {
		msg := fmt.Sprintf("Expected nil, got: %v", value)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if value == nil {
		msg := "Expected not nil"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf(msg)
	}
}

// AssertTrue asserts that a condition is true
func AssertTrue(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !condition {
		msg := "Expected true"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf(msg)
	}
}

// AssertFalse asserts that a condition is false
func AssertFalse(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if condition {
		msg := "Expected false"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf(msg)
	}
}

// AssertContains asserts that a string contains a substring
func AssertContains(t *testing.T, str, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	if !contains(str, substr) {
		msg := fmt.Sprintf("Expected '%s' to contain '%s'", str, substr)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
}

// AssertNotContains asserts that a string does not contain a substring
func AssertNotContains(t *testing.T, str, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	if contains(str, substr) {
		msg := fmt.Sprintf("Expected '%s' not to contain '%s'", str, substr)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
}

// AssertStatusCode asserts that an HTTP response has the expected status code
func AssertStatusCode(t *testing.T, expected int, resp *JSONResponse, msgAndArgs ...interface{}) {
	t.Helper()
	if resp.StatusCode != expected {
		msg := fmt.Sprintf("Expected status code %d, got %d", expected, resp.StatusCode)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
}

// AssertJSONRPCSuccess asserts that a JSON-RPC response is successful
func AssertJSONRPCSuccess(t *testing.T, resp *JSONRPCResponse, msgAndArgs ...interface{}) {
	t.Helper()
	if resp.Error != nil {
		msg := fmt.Sprintf("Expected successful JSON-RPC response, got error: %v", resp.Error)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
	if resp.Result == nil {
		msg := "Expected JSON-RPC result to be present"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf(msg)
	}
}

// AssertJSONRPCError asserts that a JSON-RPC response has an error
func AssertJSONRPCError(t *testing.T, resp *JSONRPCResponse, expectedCode int, msgAndArgs ...interface{}) {
	t.Helper()
	if resp.Error == nil {
		msg := "Expected JSON-RPC error but got success"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf(msg)
		return
	}
	if expectedCode != 0 && resp.Error.Code != expectedCode {
		msg := fmt.Sprintf("Expected error code %d, got %d", expectedCode, resp.Error.Code)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
}

// AssertMapKeyExists asserts that a map contains a specific key
func AssertMapKeyExists(t *testing.T, m map[string]interface{}, key string, msgAndArgs ...interface{}) {
	t.Helper()
	if _, exists := m[key]; !exists {
		msg := fmt.Sprintf("Expected map to contain key '%s'", key)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
}

// AssertMapKeyValue asserts that a map contains a specific key-value pair
func AssertMapKeyValue(t *testing.T, m map[string]interface{}, key string, expectedValue interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	value, exists := m[key]
	if !exists {
		msg := fmt.Sprintf("Expected map to contain key '%s'", key)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
		return
	}
	if !reflect.DeepEqual(expectedValue, value) {
		msg := fmt.Sprintf("Expected map[%s] = %v, got %v", key, expectedValue, value)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		t.Errorf(msg)
	}
}

// Helper function to check if string contains substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(substr) > 0 && indexOfSubstring(str, substr) >= 0))
}

// Helper function to find index of substring
func indexOfSubstring(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
