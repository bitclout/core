package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSliceAll(t *testing.T) {
	// Predicate: all values > 0
	predicate := func(val int) bool {
		return val > 0
	}

	// Test sad path where no values are > 0
	{
		slice := []int{-1, -2, -3, -4, -5}
		require.False(t, All(slice, predicate))
	}

	// Test sad path where some values are > 0
	{
		slice := []int{-1, 2, 3, 4, 5}
		require.False(t, All(slice, predicate))
	}

	// Test happy path where all values are > 0
	{
		slice := []int{1, 2, 3, 4, 5}
		require.True(t, All(slice, predicate))
	}
}

func TestSliceAny(t *testing.T) {
	// Predicate: all values > 0
	predicate := func(val int) bool {
		return val > 0
	}

	// Test sad path where no values are > 0
	{
		slice := []int{-1, -2, -3, -4, -5}
		require.False(t, Any(slice, predicate))
	}

	// Test happy path where some values are > 0
	{
		slice := []int{-1, 2, 3, 4, 5}
		require.True(t, Any(slice, predicate))
	}

	// Test happy path where all values are > 0
	{
		slice := []int{1, 2, 3, 4, 5}
		require.True(t, Any(slice, predicate))
	}
}

func TestSliceToMap(t *testing.T) {
	// Create a struct to test the slice -> map transformation
	type keyValueType struct {
		Key   string
		Value string
	}

	// Test empty slice
	{
		// Create a custom function extract the key from the struct
		keyFn := func(val keyValueType) string {
			return val.Key
		}

		slice := []keyValueType{}
		result := ToMap(slice, keyFn)
		require.Equal(t, 0, len(result))
	}

	// Test slice with pointers
	{
		// Create a custom function extract the key from the struct
		keyFn := func(val *keyValueType) string {
			return val.Key
		}

		slice := []*keyValueType{
			{Key: "a", Value: "1"},
			{Key: "b", Value: "2"},
		}
		result := ToMap(slice, keyFn)
		require.Equal(t, 2, len(result))
		require.Equal(t, "1", result["a"].Value)
		require.Equal(t, "2", result["b"].Value)
	}

	// Test slice with raw values
	{
		// Create a custom function extract the key from the struct
		keyFn := func(val keyValueType) string {
			return val.Key
		}

		slice := []keyValueType{
			{Key: "a", Value: "1"},
			{Key: "b", Value: "2"},
		}
		result := ToMap(slice, keyFn)
		require.Equal(t, 2, len(result))
		require.Equal(t, "1", result["a"].Value)
		require.Equal(t, "2", result["b"].Value)
	}
}
