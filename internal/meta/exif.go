package meta

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dsoprea/go-exif"
	"gopkg.in/ugjka/go-tz.v2/tz"
)

// Exif parses an image file for Exif meta data and returns as Data struct.
func Exif(filename string) (data Data, err error) {
	defer func() {
		if e := recover(); e != nil {
			data = Data{}
			err = fmt.Errorf("meta: %s", e)
		}
	}()

	// TODO: Requires refactoring and performance optimization
	// See GitHub: https://github.com/photoprism/photoprism/issues/172
	rawExif, err := exif.SearchFileAndExtractExif(filename)

	if err != nil {
		return data, err
	}

	ti := exif.NewTagIndex()

	tags := make(map[string]string)

	visitor := func(fqIfdPath string, ifdIndex int, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (err error) {
		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)

		if err != nil {
			return nil
		}

		it, err := ti.Get(ifdPath, tagId)

		if err != nil {
			return nil
		}

		valueString := ""

		if tagType.Type() != exif.TypeUndefined {
			valueString, err = tagType.ResolveAsString(valueContext, true)

			if err != nil {
				log.Error(err)

				return nil
			}

			if it.Name != "" && valueString != "" {
				tags[it.Name] = valueString
			}
		}

		return nil
	}

	_, err = exif.Visit(exif.IfdStandard, im, ti, rawExif, visitor)

	if err != nil {
		return data, err
	}

	if value, ok := tags["Artist"]; ok {
		data.Artist = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["Copyright"]; ok {
		data.Copyright = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["Model"]; ok {
		data.CameraModel = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["Make"]; ok {
		data.CameraMake = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["LensMake"]; ok {
		data.LensMake = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["LensModel"]; ok {
		data.LensModel = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["ExposureTime"]; ok {
		data.Exposure = value
	}

	if value, ok := tags["FNumber"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			data.FNumber = math.Round((number/denom)*1000) / 1000
		}
	}

	if value, ok := tags["ApertureValue"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			data.Aperture = math.Round((number/denom)*1000) / 1000
		}
	}

	if value, ok := tags["FocalLengthIn35mmFilm"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.FocalLength = i
		}
	} else if value, ok := tags["FocalLength"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			data.FocalLength = int(math.Round((number/denom)*1000) / 1000)
		}
	}

	if value, ok := tags["ISOSpeedRatings"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Iso = i
		}
	}

	if value, ok := tags["ImageUniqueID"]; ok {
		data.UUID = value
	}

	if value, ok := tags["ImageWidth"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Width = i
		}
	}

	if value, ok := tags["ImageLength"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Height = i
		}
	}

	if value, ok := tags["Orientation"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Orientation = i
		}
	} else {
		data.Orientation = 1
	}

	_, index, err := exif.Collect(im, ti, rawExif)

	if err != nil {
		return data, err
	}

	if ifd, err := index.RootIfd.ChildWithIfdPath(exif.IfdPathStandardGps); err == nil {
		if gi, err := ifd.GpsInfo(); err == nil {
			data.Lat = gi.Latitude.Decimal()
			data.Lng = gi.Longitude.Decimal()
			data.Altitude = gi.Altitude
		}
	}

	if data.Lat != 0 && data.Lng != 0 {
		zones, err := tz.GetZone(tz.Point{
			Lat: data.Lat,
			Lon: data.Lng,
		})

		if err != nil {
			data.TimeZone = "UTC"
		}

		data.TimeZone = zones[0]
	}

	if value, ok := tags["DateTimeOriginal"]; ok {
		data.TakenAtLocal, _ = time.Parse("2006:01:02 15:04:05", value)

		loc, err := time.LoadLocation(data.TimeZone)

		if err != nil {
			data.TakenAt = data.TakenAtLocal
			log.Warnf("no location for timezone: %s", err.Error())
		} else if tl, err := time.ParseInLocation("2006:01:02 15:04:05", value, loc); err == nil {
			data.TakenAt = tl.UTC()
		} else {
			log.Warnf("could parse time: %s", err.Error())
		}
	}

	if value, ok := tags["Flash"]; ok {
		if i, err := strconv.Atoi(value); err == nil && i&1 == 1 {
			data.Flash = true
		}
	}

	if value, ok := tags["ImageDescription"]; ok {
		data.Description = strings.Replace(value, "\"", "", -1)
	}

	data.All = tags

	return data, nil
}
