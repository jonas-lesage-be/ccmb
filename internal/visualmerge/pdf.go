package visualmerge

import (
	"context"
	"errors"
	"fmt"
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

func (m *Merger) preparePDFComponent(ctx context.Context, j job) (string, int64, error) {
	convertedPDF, err := m.convertToHeaderedPDF(ctx, j)
	if err != nil {
		return "", 0, err
	}

	fi, err := os.Stat(convertedPDF)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			if err := os.Remove(convertedPDF); err != nil {
				slog.Error("Failed to delete file", "path", convertedPDF, "err", err)
			}
		}
		return "", 0, fmt.Errorf("failed to check PDF status: %w", err)
	}

	return convertedPDF, fi.Size(), nil
}

func (m *Merger) convertToHeaderedPDF(ctx context.Context, j job) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context error before converting file %s: %w", j.path, err)
	}

	ext := strings.ToLower(filepath.Ext(j.name))
	outPDF := filepath.Join(os.TempDir(), fmt.Sprintf("ccmb_tmp_%d.pdf", j.index))

	wm, err := m.createWatermark(j.name, ext)
	if err != nil {
		return "", fmt.Errorf("failed to create watermark for %s: %w", j.name, err)
	}

	if ext == ".pdf" {
		err = pdfcpu_api.AddWatermarksFile(j.path, outPDF, nil, wm, nil)
		if err != nil {
			return "", fmt.Errorf("failed to add watermark to PDF %s: %w", j.name, err)
		}
	} else {
		err = importImageWithWatermark(j.path, outPDF, j.name, wm)
		if err != nil {
			return "", fmt.Errorf("failed to import image with watermark: %w", err)
		}
	}

	return outPDF, nil
}

func (m *Merger) createWatermark(flatName, ext string) (*pdfcpu_model.Watermark, error) {
	desc := "pos: tr, off: -6 -6, points: 10, scale: 1.0 abs, rot: 0, mode: 0," +
		" color: #000000, bgcol: #ffffff, border: 1 #000000, margins: 4"
	wm, err := pdfcpu.ParseTextWatermarkDetails(
		m.headerText(flatName, ext),
		desc,
		true,
		pdfcpu_types.POINTS,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pdfcpu watermark details: %w", err)
	}

	return wm, nil
}

func (m *Merger) mergeBatch(files []string, counter int) error {
	fname := fmt.Sprintf("FinalResult_Visual_Part_%d.pdf", counter)
	fpath := filepath.Join(m.TargetDir, fname)
	slog.Info("Flushing and writing structured batch out to file payload", "file", fname)

	err := pdfcpu_api.MergeCreateFile(files, fpath, false, nil)
	if err != nil {
		return fmt.Errorf("failed to merge batch into %s: %w", fname, err)
	}

	if err := cleanUpFiles(files); err != nil {
		return fmt.Errorf("failed to clean up temporary batch files: %w", err)
	}

	return nil
}

func (m *Merger) headerText(flatName, ext string) string {
	if strings.Contains(flatName, m.FlatPathDelimiter+"frame_") {
		before, after, _ := strings.Cut(flatName, m.FlatPathDelimiter+"frame_")
		frameNum := strings.TrimSuffix(after, ext)

		if _, err := strconv.Atoi(frameNum); err == nil {
			cleanPath := m.originalPathFromFlatName(before)
			return fmt.Sprintf("File path: %s frame number %s", cleanPath, frameNum)
		}
	}

	targetSuffix := m.FlatPathDelimiter + ".png"
	if before, ok := strings.CutSuffix(flatName, targetSuffix); before != "" && ok {
		originalPath := m.originalPathFromFlatName(before)
		return "File path: " + originalPath
	}

	originalPath := m.originalPathFromFlatName(flatName)
	return "File path: " + originalPath
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
