package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLV(t *testing.T) {

	testExamples := []struct {
		templateLine string
		exceptResult *LV
		exceptError  error
	}{
		{
			templateLine: "vgs --units=b xxxx --separator='<:SEP:>' --nosuffix --noheadings -o vg_name,vg_size,vg_free,vg_uuid,vg_tags --nameprefixes -a",
			exceptResult: &LV{},
			exceptError:  errors.New("expected 8 components, got 2"),
		},
		{
			templateLine: "vgs --units=b --separator='<:SEP:>' --nosuffix --noheadings -o vg_name,vg_size,vg_free,vg_uuid,vg_tags --nameprefixes -a",
			exceptResult: &LV{},
			exceptError:  errors.New("expected 8 components, got 2"),
		},
	}

	for _, e := range testExamples {
		_, err := ParseLV(e.templateLine)
		assert.EqualError(t, e.exceptError, err.Error())

	}
}
