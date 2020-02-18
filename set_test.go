package mtgdb_test

import (
	"testing"

	"github.com/pioz/mtgdb"
	"github.com/stretchr/testify/assert"
)

func TestSetImagePath(t *testing.T) {
	set := mtgdb.Set{
		Code:     "peld",
		IconName: "eld",
	}
	assert.Equal(t, "images/sets/eld.jpg", set.ImagePath("./images"))
}
