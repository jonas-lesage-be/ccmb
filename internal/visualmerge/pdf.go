package visualmerge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	pdfcpu_api "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	pdfcpu_types "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"ccmb/internal/config"
)

func (m *Merger) preparePDFComponent(ctx context.Context, j job) (jobResult, error) {
	path, data, err := m.convertToHeaderedPDF(ctx, j)
	if err != nil {
		return jobResult{}, err
	}

	if m.InMemory {
		return jobResult{path: "", data: data, size: int64(len(data))}, nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			if err := os.Remove(path); err != nil {
				slog.Error("Failed to delete file", "path", path, "err", err)
			}
		}
		return jobResult{}, fmt.Errorf("failed to check PDF status: %w", err)
	}

	return jobResult{path: path, data: nil, size: fi.Size()}, nil
}

func (m *Merger) convertToHeaderedPDF(ctx context.Context, j job) (string, []byte, error) {
	if err := ctx.Err(); err != nil {
		return "", nil, fmt.Errorf("context error before converting file %s: %w", j.path, err)
	}

	wm, err := m.createWatermark(j)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create watermark for %s: %w", j.name, err)
	}

	if m.InMemory {
		data, err := convertToHeaderedPDFInMemory(j, wm)
		if err != nil {
			return "", nil, err
		}
		return "", data, nil
	}

	path, err := convertToHeaderedPDFOnDisk(j, wm)
	if err != nil {
		return "", nil, err
	}
	return path, nil, nil
}

func convertToHeaderedPDFInMemory(
	j job,
	wm *pdfcpu_model.Watermark,
) ([]byte, error) {
	data, err := convertToPDFInMemory(j)
	if err != nil {
		return nil, err
	}

	var watermarkedPDF bytes.Buffer
	if err := pdfcpu_api.AddWatermarks(data, &watermarkedPDF, nil, wm, nil); err != nil {
		return nil, fmt.Errorf("failed to add watermark to %s: %w", j.name, err)
	}

	return watermarkedPDF.Bytes(), nil
}

func convertToPDFInMemory(j job) (*bytes.Reader, error) {
	cleanSrcPath := filepath.Clean(j.path)

	if j.ext == ".pdf" {
		pdf, err := os.ReadFile(cleanSrcPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read PDF %s: %w", j.name, err)
		}
		return bytes.NewReader(pdf), nil
	}

	imageFile, err := os.Open(cleanSrcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image %s: %w", j.name, err)
	}
	defer func() {
		if err := imageFile.Close(); err != nil {
			slog.Error("Failed to close file", "path", cleanSrcPath, "err", err)
		}
	}()

	var importedPDF bytes.Buffer
	imp := pdfcpu.DefaultImportConfig()
	imp.PageSize = "Letter"
	if err := pdfcpu_api.ImportImages(
		nil,
		&importedPDF,
		[]io.Reader{imageFile},
		imp,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to import image %s to PDF: %w", j.name, err)
	}

	return bytes.NewReader(importedPDF.Bytes()), nil
}

func convertToHeaderedPDFOnDisk(
	j job,
	wm *pdfcpu_model.Watermark,
) (string, error) {
	outPDF := filepath.Join(os.TempDir(), fmt.Sprintf("ccmb_tmp_%d.pdf", j.index))

	if j.ext == ".pdf" {
		if err := pdfcpu_api.AddWatermarksFile(j.path, outPDF, nil, wm, nil); err != nil {
			return "", fmt.Errorf("failed to add watermark to PDF %s: %w", j.name, err)
		}
	} else {
		if err := importImageWithWatermark(j.path, outPDF, j.name, wm); err != nil {
			return "", fmt.Errorf("failed to import image with watermark: %w", err)
		}
	}

	return outPDF, nil
}

func (m *Merger) createWatermark(j job) (*pdfcpu_model.Watermark, error) {
	desc := "pos: tr, off: -6 -6, points: 10, scale: 1.0 abs, rot: 0, mode: 0," +
		" color: #000000, bgcol: #ffffff, border: 1 #000000, margins: 4"
	wm, err := pdfcpu.ParseTextWatermarkDetails(
		m.headerText(j),
		desc,
		true,
		pdfcpu_types.POINTS,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pdfcpu watermark details: %w", err)
	}

	return wm, nil
}

func (m *Merger) mergeBatch(batch []jobResult, counter int) error {
	fname := fmt.Sprintf("%s%d.pdf", m.VisualPartPrefix, counter)
	fpath := filepath.Clean(filepath.Join(m.TargetDir, fname))
	slog.Info("Flushing and writing structured batch out to file payload", "file", fname)

	var err error
	if m.InMemory {
		err = mergeBatchInMemory(fname, fpath, batch)
	} else {
		err = mergeBatchOnDisk(fname, fpath, batch)
	}

	if err != nil {
		return fmt.Errorf("failed to merge batch: %w", err)
	}

	return nil
}

func mergeBatchInMemory(fname, fpath string, batch []jobResult) error {
	readers := make([]io.ReadSeeker, len(batch))
	for i, component := range batch {
		readers[i] = bytes.NewReader(component.data)
	}

	out, err := os.Create(filepath.Clean(fpath))
	if err != nil {
		return fmt.Errorf("failed to create merged batch %s: %w", fname, err)
	}
	defer func() {
		if err := out.Close(); err != nil {
			slog.Error("Failed to close file", "path", fpath, "err", err)
		}
	}()

	if err := pdfcpu_api.MergeRaw(readers, out, false, nil); err != nil {
		return fmt.Errorf("failed to merge batch into %s: %w", fname, err)
	}

	return nil
}

func mergeBatchOnDisk(fname, fpath string, batch []jobResult) error {
	files := make([]string, len(batch))
	for i, component := range batch {
		files[i] = component.path
	}

	if err := pdfcpu_api.MergeCreateFile(files, fpath, false, nil); err != nil {
		return fmt.Errorf("failed to merge batch into %s: %w", fname, err)
	}

	if err := cleanUpFiles(files); err != nil {
		return fmt.Errorf("failed to clean up temporary batch files: %w", err)
	}

	return nil
}

func (m *Merger) headerText(j job) string {
	frameMarker := m.FlatPathDelimiter + "frame_"
	if before, after, found := strings.Cut(j.name, frameMarker); found {
		frameNum := strings.TrimSuffix(after, j.ext)
		if _, err := strconv.Atoi(frameNum); err == nil {
			cleanPath := m.originalPathFromFlatName(before)
			return fmt.Sprintf("File path: %s frame number %s", cleanPath, frameNum)
		}
	}

	cleanFlatName := j.name
	if lastIdx := strings.LastIndex(j.name, m.FlatPathDelimiter); lastIdx != -1 {
		lastSegment := j.name[lastIdx+len(m.FlatPathDelimiter):]
		if strings.HasPrefix(lastSegment, ".") {
			cleanFlatName = j.name[:lastIdx]
		}
	}

	return "File path: " + m.originalPathFromFlatName(cleanFlatName)
}

func (m *Merger) originalPathFromFlatName(flatName string) string {
	return config.DecodeFlatName(
		flatName,
		m.EscapedDelimiter,
		m.DecodePlaceholder,
		m.FlatPathDelimiter,
	)
}

func importImageWithWatermark(srcPath, outPDF, flatName string, wm *pdfcpu_model.Watermark) error {
	imp := pdfcpu.DefaultImportConfig()
	imp.PageSize = "Letter"

	cleanTempFile := outPDF + ".raw.pdf"
	defer func() {
		if err := os.Remove(cleanTempFile); err != nil {
			slog.Error("Failed to delete file", "path", cleanTempFile, "err", err)
		}
	}()

	if err := pdfcpu_api.ImportImagesFile([]string{srcPath}, cleanTempFile, imp, nil); err != nil {
		return fmt.Errorf("failed to import image %s to PDF: %w", flatName, err)
	}

	if err := pdfcpu_api.AddWatermarksFile(cleanTempFile, outPDF, nil, wm, nil); err != nil {
		return fmt.Errorf("failed to add watermark to imported image PDF %s: %w", flatName, err)
	}

	return nil
}
