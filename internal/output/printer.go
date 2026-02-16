package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Printer struct {
	out    io.Writer
	asJSON bool
}

func New(out io.Writer, asJSON bool) Printer {
	return Printer{out: out, asJSON: asJSON}
}

func (p Printer) JSON(data any) error {
	enc := json.NewEncoder(p.out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (p Printer) Table(headers []string, rows [][]string) error {
	if len(headers) == 0 {
		return nil
	}

	w := tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, strings.Join(headers, "\t")); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(w, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return w.Flush()
}

func (p Printer) KV(keyVals [][2]string) error {
	w := tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
	for _, kv := range keyVals {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", kv[0], kv[1]); err != nil {
			return err
		}
	}
	return w.Flush()
}

func (p Printer) IsJSON() bool {
	return p.asJSON
}
