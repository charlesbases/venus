package spankbang

import "github.com/charlesbases/venus/regexp"

var compileParseStreamDataFromVideoLink = regexp.New(`var stream_data = (.*);`)
