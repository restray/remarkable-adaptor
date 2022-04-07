package remarkableadaptor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tablet := new(ReMarkable)

	tablet, err := tablet.Load("10.11.99.1")

	assert.NoError(t, err)
	assert.NotNil(t, tablet)
	assert.IsType(t, *tablet, ReMarkable{})
}

func TestCantLoad(t *testing.T) {
	tablet := new(ReMarkable)

	tablet, err := tablet.Load("unexisting")

	assert.Error(t, err)
	assert.Nil(t, tablet)
}

func TestFetchRootDocuments(t *testing.T) {
	tablet := new(ReMarkable)

	tablet, err := tablet.Load("10.11.99.1")

	assert.NoError(t, err)
	assert.Greater(t, len(tablet.Documents), 0)
}
