package pages

import "testing"

func TestAddFuncRegistration(t *testing.T) {
	// Arrange: Define a test function
	testFn := func() string { return "test-output" }

	// Act: Register it dynamically
	AddFunc("myMockUtility", testFn)

	// Assert: Extract the current Map and verify it exists
	currentMap := GetFuncMap()
	if _, exists := currentMap["myMockUtility"]; !exists {
		t.Error("Expected 'myMockUtility' to be registered in global map")
	}
}