package nsfw

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var modelPath, _ = filepath.Abs("../../assets/resources/nsfw")

var detector = New(modelPath)

func TestNSFW(t *testing.T) {
	detect := func(filename string) Labels {
		result, err := detector.File(filename)

		if err != nil {
			t.Fatalf(err.Error())
		}

		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)

		return result
	}

	expected := map[string]Labels{
		"beach_sand.jpg":        {0, 0, 0.9, 0, 0},
		"beach_wood.jpg":        {0, 0, 0.36, 0.59, 0},
		"cat_brown.jpg":         {0, 0, 0.93, 0, 0},
		"cat_yellow_grey.jpg":   {0, 0, 0, 0, 0.01},
		"clock_purple.jpg":      {0.19, 0, 0.80, 0, 0},
		"clowns_colorful.jpg":   {0.06, 0.02, 0.89, 0.01, 0},
		"dog.jpg":               {0.86, 0, 0.12, 0, 0},
		"hentai_1.jpg":          {0.15, 0.84, 0, 0, 0},
		"hentai_2.jpg":          {0, 0.98, 0, 0, 0},
		"hentai_3.jpg":          {0, 0.99, 0, 0, 0},
		"hentai_4.jpg":          {0, 0.94, 0, 0.05, 0},
		"hentai_5.jpg":          {0, 0.85, 0, 0.07, 0},
		"jellyfish_blue.jpg":    {0.29, 0.09, 0.57, 0, 0},
		"limes.jpg":             {0, 0.21, 0.78, 0, 0},
		"ocean_cyan.jpg":        {0, 0, 0.95, 0.03, 0},
		"peacock_blue.jpg":      {0.05, 0.05, 0.49, 0.37, 0},
		"porn_1.jpg":            {0, 0, 0, 0.97, 0},
		"porn_2.jpg":            {0, 0, 0.12, 0.77, 0},
		"porn_3.jpg":            {0, 0, 0, 0.55, 0.41},
		"porn_4.jpg":            {0, 0, 0, 0.99, 0},
		"porn_5.jpg":            {0, 0, 0.11, 0.41, 0.43},
		"porn_6.jpg":            {0, 0.1, 0.04, 0.22, 0.60},
		"porn_7.jpg":            {0, 0.25, 0, 0.66, 0},
		"porn_8.jpg":            {0, 0.12, 0, 0.86, 0.01},
		"porn_9.jpg":            {0.95, 0.02, 0, 0.01, 0},
		"porn_10.jpg":           {0, 0.05, 0, 0.79, 0.13},
		"porn_11.jpg":           {0, 0, 0.09, 0.36, 0.53},
		"sexy_1.jpg":            {0.02, 0.49, 0.01, 0, 0.46},
		"sharks_blue.jpg":       {0.22, 0.007, 0.75, 0, 0},
		"zebra_green_brown.jpg": {0.24, 0.01, 0.73, 0.004, 0.001},
	}

	err := filepath.Walk("testdata", func(filename string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		t.Run(filename, func(t *testing.T) {
			l := detect(filename)

			basename := filepath.Base(filename)

			t.Logf("labels:   %+v", l)

			if e, ok := expected[basename]; ok {
				t.Logf("expected: %+v", e)

				assert.GreaterOrEqual(t, l.Drawing, e.Drawing)
				assert.GreaterOrEqual(t, l.Hentai, e.Hentai)
				assert.GreaterOrEqual(t, l.Neutral, e.Neutral)
				assert.GreaterOrEqual(t, l.Porn, e.Porn)
				assert.GreaterOrEqual(t, l.Sexy, e.Sexy)
			}

			isSafe := !(strings.Contains(basename, "porn") || strings.Contains(basename, "hentai"))

			if isSafe {
				assert.True(t, l.IsSafe())
			}
		})

		return nil
	})

	if err != nil {
		t.Log(err.Error())
	}
}
