/*
	Copyright 2020 The pdfcpu Authors.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package api

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pirogom/pdfcpu/pkg/log"
	"github.com/pirogom/pdfcpu/pkg/pdfcpu"
)

type PageSpan struct {
	From   int
	Thru   int
	Reader io.Reader
}

func pageSpan(ctx *pdfcpu.Context, from, thru int) (*PageSpan, error) {
	ctxNew, err := ctx.ExtractPages(PagesForPageRange(from, thru), false)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	if err := WriteContext(ctxNew, &b); err != nil {
		return nil, err
	}

	return &PageSpan{From: from, Thru: thru, Reader: &b}, nil
}

func spanFileName(fileName string, from, thru int) string {
	baseFileName := filepath.Base(fileName)
	fn := strings.TrimSuffix(baseFileName, ".pdf")
	fn = fn + "_" + strconv.Itoa(from)
	if from == thru {
		return fn + ".pdf"
	}
	return fn + "-" + strconv.Itoa(thru) + ".pdf"
}

func splitOutPath(outDir, fileName string, forBookmark bool, from, thru int) string {
	p := filepath.Join(outDir, fileName+".pdf")
	if !forBookmark {
		p = filepath.Join(outDir, spanFileName(fileName, from, thru))
	}
	return p
}

func writePageSpan(ctx *pdfcpu.Context, from, thru int, outPath string) error {
	ps, err := pageSpan(ctx, from, thru)
	if err != nil {
		return err
	}
	if log.CLI != nil {
		log.CLI.Printf("writing %s...\n", outPath)
	}
	return pdfcpu.WriteReader(outPath, ps.Reader)
}

func context(rs io.ReadSeeker, conf *pdfcpu.Configuration) (*pdfcpu.Context, error) {
	if conf == nil {
		conf = pdfcpu.NewDefaultConfiguration()
	}
	conf.Cmd = pdfcpu.SPLIT

	ctx, _, _, _, err := readValidateAndOptimize(rs, conf, time.Now())
	if err != nil {
		return nil, err
	}

	if err := ctx.EnsurePageCount(); err != nil {
		return nil, err
	}

	return ctx, nil
}

func pageSpansSplitAlongBookmarks(ctx *pdfcpu.Context) ([]*PageSpan, error) {
	pss := []*PageSpan{}

	bms, err := ctx.BookmarksForOutline()
	if err != nil {
		return nil, err
	}

	for _, bm := range bms {

		from, thru := bm.PageFrom, bm.PageThru
		if thru == 0 {
			thru = ctx.PageCount
		}

		ps, err := pageSpan(ctx, from, thru)
		if err != nil {
			return nil, err
		}
		pss = append(pss, ps)

	}

	return pss, nil
}

func pageSpans(ctx *pdfcpu.Context, span int) ([]*PageSpan, error) {
	pss := []*PageSpan{}

	for i := 0; i < ctx.PageCount/span; i++ {
		start := i * span
		from := start + 1
		thru := start + span
		ps, err := pageSpan(ctx, from, thru)
		if err != nil {
			return nil, err
		}
		pss = append(pss, ps)
	}

	// A possible last file has less than span pages.
	if ctx.PageCount%span > 0 {
		start := (ctx.PageCount / span) * span
		from := start + 1
		thru := ctx.PageCount
		ps, err := pageSpan(ctx, from, thru)
		if err != nil {
			return nil, err
		}
		pss = append(pss, ps)
	}

	return pss, nil
}

func writePageSpansSplitAlongBookmarks(ctx *pdfcpu.Context, outDir string) error {
	forBookmark := true

	bms, err := ctx.BookmarksForOutline()
	if err != nil {
		return err
	}

	for _, bm := range bms {
		fileName := strings.Replace(bm.Title, " ", "_", -1)
		from, thru := bm.PageFrom, bm.PageThru
		if thru == 0 {
			thru = ctx.PageCount
		}
		path := splitOutPath(outDir, fileName, forBookmark, from, thru)
		if err := writePageSpan(ctx, from, thru, path); err != nil {
			return err
		}
	}

	return nil
}

func writePageSpans(ctx *pdfcpu.Context, span int, outDir, fileName string) error {
	forBookmark := false

	for i := 0; i < ctx.PageCount/span; i++ {
		start := i * span
		from, thru := start+1, start+span
		path := splitOutPath(outDir, fileName, forBookmark, from, thru)
		if err := writePageSpan(ctx, from, thru, path); err != nil {
			return err
		}
	}

	// A possible last file has less than span pages.
	if ctx.PageCount%span > 0 {
		start := (ctx.PageCount / span) * span
		from, thru := start+1, ctx.PageCount
		path := splitOutPath(outDir, fileName, forBookmark, from, thru)
		if err := writePageSpan(ctx, from, thru, path); err != nil {
			return err
		}
	}

	return nil
}

// SplitRaw returns page spans for the PDF stream read from rs obeying given split span.
// If span == 1 splitting results in single page PDFs.
// If span == 0 we split along given bookmarks (level 1 only).
// Default span: 1
func SplitRaw(rs io.ReadSeeker, span int, conf *pdfcpu.Configuration) ([]*PageSpan, error) {
	ctx, err := context(rs, conf)
	if err != nil {
		return nil, err
	}

	if span == 0 {
		return pageSpansSplitAlongBookmarks(ctx)
	}
	return pageSpans(ctx, span)
}

// Split generates a sequence of PDF files in outDir for the PDF stream read from rs obeying given split span.
// If span == 1 splitting results in single page PDFs.
// If span == 0 we split along given bookmarks (level 1 only).
// Default span: 1
func Split(rs io.ReadSeeker, outDir, fileName string, span int, conf *pdfcpu.Configuration) error {
	ctx, err := context(rs, conf)
	if err != nil {
		return err
	}

	if span == 0 {
		return writePageSpansSplitAlongBookmarks(ctx, outDir)
	}
	return writePageSpans(ctx, span, outDir, fileName)
}

// SplitFile generates a sequence of PDF files in outDir for inFile obeying given split span.
// If span == 1 splitting results in single page PDFs.
// If span == 0 we split along given bookmarks (level 1 only).
// Default span: 1
func SplitFile(inFile, outDir string, span int, conf *pdfcpu.Configuration) error {
	f, err := os.Open(inFile)
	if err != nil {
		return err
	}
	log.CLI.Printf("splitting %s to %s/...\n", inFile, outDir)

	defer func() {
		if err != nil {
			f.Close()
			return
		}
		err = f.Close()
	}()

	return Split(f, outDir, filepath.Base(inFile), span, conf)
}

// Get Bookmark info
func GetBookmark(inFile string, conf *pdfcpu.Configuration) ([]pdfcpu.Bookmark, error) {
	f, err := os.Open(inFile)

	if err != nil {
		return nil, err
	}
	log.CLI.Printf("get bookmark data from %s", inFile)

	defer f.Close()

	ctx, ctxErr := context(f, conf)

	if ctxErr != nil {
		return nil, ctxErr
	}

	bms, bmsErr := ctx.BookmarksForOutline()
	if bmsErr != nil {
		return nil, bmsErr
	}
	return bms, nil
}