package printer

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dty1er/kubecolor/color"
)

type DescribePrinter struct {
	Writer io.Writer
}

func (dp *DescribePrinter) Print(outReader io.Reader) {
	w := dp.Writer
	scanner := bufio.NewScanner(outReader)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			fmt.Fprintln(w)
			continue
		}

		columns := tab.Split(line, -1)
		if len(columns) > 0 {
			if columns[0] == "" {
				columns = columns[1:]
			}
		}

		if len(columns) > 0 {
			if columns[len(columns)-1] == "" {
				columns = columns[:len(columns)-1]
			}
		}

		spacesIndices := tab.FindAllStringIndex(line, -1)

		if len(columns) == 1 {
			// when coming here, the format is not "key: value" but only key: or only value
			if spacesIndices == nil {
				// no indent
				fmt.Fprintf(w, "%s\n", dp.colorizeKeyOrValue(columns[0], 0))
				continue
			}

			// spacesIndices[0] represents the position of indent spaces
			indentPos := spacesIndices[0]
			indent := indentPos[1] - indentPos[0]
			fmt.Fprintf(w, "%s%s\n", dp.toSpaces(indent), dp.colorizeKeyOrValue(columns[0], indent))
			continue
		}

		// When coming here, the line must have key and value; but value might be multiple

		if len(columns) != len(spacesIndices) {
			// when coming here, indent must not exist
			// First, write key
			fmt.Fprintf(w, "%s", dp.colorizeKeyOrValue(columns[0], 0))
			// Then, write values
			for i, column := range columns {
				if i == 0 {
					continue
				}

				spacesPos := spacesIndices[i-1]
				spaceCnt := spacesPos[1] - spacesPos[0]
				c := dp.colorByValue(column)
				fmt.Fprintf(w, "%s%s", dp.toSpaces(spaceCnt), color.Apply(column, c))
			}

			fmt.Fprintf(w, "\n")
			continue
		}

		// when coming here, indent must exists
		// First, write key
		indentPos := spacesIndices[0]
		indent := indentPos[1] - indentPos[0]
		fmt.Fprintf(w, "%s%s", dp.toSpaces(indent), dp.colorizeKeyOrValue(columns[0], indent))
		// Then, write values
		for i, column := range columns {
			if i == 0 {
				continue
			}

			spacesPos := spacesIndices[i]
			spacesCnt := spacesPos[1] - spacesPos[0]
			c := dp.colorByValue(column)
			fmt.Fprintf(w, "%s%s", dp.toSpaces(spacesCnt), color.Apply(column, c))
		}

		fmt.Fprint(w, "\n")
	}
}

func (dp *DescribePrinter) colorizeKeyOrValue(s string, indent int) string {
	if strings.HasSuffix(s, ":") {
		ss := strings.TrimRight(s, ":") // trailing ":" should not be colorized
		return color.Apply(ss, dp.colorByIndent(indent)) + ":"
	}

	return color.Apply(s, dp.colorByValue(s))
}

func (dp *DescribePrinter) toSpaces(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += " "
	}
	return s
}

func (dp *DescribePrinter) colorByIndent(indent int) color.Color {
	switch indent / 2 % 2 {
	case 1:
		return color.Blue
	default:
		return color.Yellow
	}
}

func (dp *DescribePrinter) colorByValue(val string) color.Color {
	if val == "<none>" || val == "<unknown>" {
		return NullColor
	}

	if val == "true" || val == "false" {
		return BoolColor
	}

	if _, err := strconv.Atoi(val); err == nil {
		return NumberColor
	}

	return StringColor
}
