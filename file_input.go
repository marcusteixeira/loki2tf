// https://github.com/sl1pm4t/k2tf/blob/master/pkg/file_io/input.go
// Adapted to fit to mimir promehteus rules files.

package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/mimir/pkg/mimirtool/rules"

	"github.com/rs/zerolog/log"
)

func ReadInput(input string) []rules.RuleNamespace {
	if input == "-" || input == "" {
		return readStdinInput(input)
	}
	return readFilesInput(input)
}

func readStdinInput(input string) []rules.RuleNamespace {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if info.Mode()&os.ModeCharDevice != 0 {
		log.Fatal().Msg("No data read from stdin")
	}

	reader := bufio.NewReader(os.Stdin)
	buf := &bytes.Buffer{}
	buf.ReadFrom(reader)
	parsed, errs := rules.ParseBytes(buf.Bytes())

	if len(errs) > 0 {
		log.Fatal().Err(errs[0]).Msg("Could not parse stdin")
	}

	return parsed

}

func readFilesInput(input string) []rules.RuleNamespace {
	var objs []rules.RuleNamespace

	if _, err := os.Stat(input); os.IsNotExist(err) {
		log.Fatal().Str("file", input).Msg("input filepath does not exist")
	}

	file, err := os.Open(input)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	fs, err := file.Stat()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	readFile := func(fileName string) {
		log.Debug().Msgf("reading file: %s", fileName)
		content, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Fatal().Err(err).Msg("could not read file")
		}

		r := bytes.NewReader(content)
		buf := &bytes.Buffer{}
		buf.ReadFrom(r)
		obj, errs := rules.ParseBytes(buf.Bytes())
		if len(errs) > 0 {
			log.Warn().Err(errs[0]).Msgf("could not parse file %s", fileName)
		}
		objs = append(objs, obj...)
	}

	if fs.Mode().IsDir() {
		// read directory
		log.Debug().Msgf("reading directory: %s", input)

		dirContents, err := file.Readdirnames(0)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}

		for _, f := range dirContents {
			if strings.HasSuffix(f, ".yml") || strings.HasSuffix(f, ".yaml") {
				readFile(filepath.Join(input, f))
			}
		}

	} else {
		// read single file
		readFile(input)

	}

	return objs
}
