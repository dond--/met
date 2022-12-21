package exif

import (
  "fmt"
  "github.com/dsoprea/go-exif/v3"
  "strconv"
  "strings"
  "time"
)

func GetTime(extime string) (time.Time, error) {
  var t time.Time
  datetime := strings.Split(extime, " ")
  if(len(datetime) != 3) {
    return t, fmt.Errorf("Can't parse EXIF datetime: " + extime)
    // we're expecting: YYYY:MM:DD hh:mm:ss +th:tm (th = timezone hour offset, tm = timezone minute offset)
  } else {
    ymd := strings.Split(datetime[0], ":")
    hms := strings.Split(datetime[1], ":")
//     offset := strings.Split(datetime[2], ":")
    Y, err := strconv.Atoi(ymd[0]) // YEAR
    if err != nil { return t, err }
    M, err := strconv.Atoi(ymd[1]) // MONTH #
    if err != nil { return t, err }
    D, err := strconv.Atoi(ymd[2]) // DAY
    if err != nil { return t, err }
    h, err := strconv.Atoi(hms[0]) // hour
    if err != nil { return t, err }
    m, err := strconv.Atoi(hms[1]) // minute
    if err != nil { return t, err }
    s, err := strconv.Atoi(hms[2]) // sec
    if err != nil { return t, err }
//     tz, err := strconv.Atoi(offset[0]) // "timezone"
//     if err != nil { return t, err }
//     t = time.Date(Y, time.Month(M), D, h, m, s, 0, time.FixedZone("EXIF", tz*60*60 ))
    t = time.Date(Y, time.Month(M), D, h, m, s, 0, time.Local)
  }
  return t, nil
}

func ReadExifTime(file string) (string, error) {
  data, err := exif.SearchFileAndExtractExif(file)
  if err != nil {
    return "", err
  }
	tags, _, err := exif.GetFlatExifDataUniversalSearch(data, nil, true)
  if err != nil {
    return "", err
  }
  var datetime string
  var timezone string
  for _, t := range tags {
    if t.TagName == "DateTimeOriginal" {
      datetime = t.Formatted
    }
    if t.TagName == "OffsetTimeOriginal" {
      timezone = t.Formatted
    }
//     fmt.Printf("IFD-PATH=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n", t.IfdPath, t.TagId, t.TagName, t.UnitCount, t.TagTypeName, t.Formatted)
  }
  if datetime == "" {
    return "", nil
  }
  if timezone == "" {
    timezone = "+00:00" // UTC equiv
  }
  return datetime + " " + timezone, nil
}
