package wire

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFoodGroupName_HappyPath(t *testing.T) {
	assert.Equal(t, "OService", FoodGroupName(OService))
}

func TestFoodGroupName_InvalidFoodGroup(t *testing.T) {
	assert.Equal(t, "unknown", FoodGroupName(2142))
}

func TestSubGroupName_HappyPath(t *testing.T) {
	assert.Equal(t, "OServiceServiceRequest", SubGroupName(OService, OServiceServiceRequest))
}

func TestSubGroupName_InvalidFoodGroup(t *testing.T) {
	assert.Equal(t, "unknown", SubGroupName(2142, OServiceServiceRequest))
}
