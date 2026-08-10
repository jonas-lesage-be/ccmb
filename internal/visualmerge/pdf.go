package visualmerge

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	pdfcpu_api "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	pdfcpu_types "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"ccmb/internal/config"
)

func (m *Merger) preparePDFComponent(ctx context.Context, fullPath string) (string, int64, error) {
	convertedPDF, err := m.convertToHeaderedPDF(ctx, fullPath, m.TargetDir)
	if err != nil {
		return "", 0, err
	}

	fi, err := os.Stat(convertedPDF)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			if rmErr := os.Remove(convertedPDF); rmErr != nil {
				log.Printf("Failed to delete %s: %v", convertedPDF, rmErr)
			}
		}
		return "", 0, fmt.Errorf("failed to check PDF status: %w", err)
	}

	return convertedPDF, fi.Size(), nil
}

func (m *Merger) convertToHeaderedPDF(
	ctx context.Context,
	srcPath, targetDir string,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context error before converting file %s: %w", srcPath, err)
	}

	flatName := filepath.Base(srcPath)
	ext := strings.ToLower(filepath.Ext(flatName))
	outPDF := filepath.Join(targetDir, "tmp_"+flatName+".pdf")

	wm, err := m.createWatermark(flatName, ext)
	if err != nil {
		return "", fmt.Errorf("failed to create watermark for %s: %w", flatName, err)
	}

	if ext == ".pdf" {
		err = pdfcpu_api.AddWatermarksFile(srcPath, outPDF, nil, wm, nil)
		if err != nil {
			return "", fmt.Errorf("failed to add watermark to PDF %s: %w", flatName, err)
		}
	} else {
		err = importImageWithWatermark(srcPath, outPDF, flatName, wm)
		if err != nil {
			return "", fmt.Errorf("failed to import image with watermark: %w", err)
		}
	}

	return outPDF, nil
}

func (m *Merger) createWatermark(flatName, ext string) (*pdfcpu_model.Watermark, error) {
	originalPath := config.DecodeFlatName(
		flatName,
		m.EscapedDelimiter,
		m.DecodePlaceholder,
		m.FlatPathDelimiter,
	)
	headerText := "File path: " + originalPath

	if strings.Contains(flatName, m.FlatPathDelimiter+"frame_") {
		before, after, _ := strings.Cut(flatName, m.FlatPathDelimiter+"frame_")
		cleanPath := config.DecodeFlatName(
			before,
			m.EscapedDelimiter,
			m.DecodePlaceholder,
			m.FlatPathDelimiter,
		)
		frameNum := strings.TrimSuffix(after, ext)
		headerText = fmt.Sprintf("File path: %s frame number %s", cleanPath, frameNum)
	}

	desc := "pos: tr, off: -6 -6, points: 10, scale: 1.0 abs, rot: 0, mode: 0," +
		" color: #000000, bgcol: #ffffff, border: 1 #000000, margins: 4"
	wm, err := pdfcpu.ParseTextWatermarkDetails(
		headerText,
		desc,
		true,
		pdfcpu_types.POINTS,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pdfcpu watermark details: %w", err)
	}

	return wm, nil
}

func importImageWithWatermark(srcPath, outPDF, flatName string, wm *pdfcpu_model.Watermark) error {
	imp := pdfcpu.DefaultImportConfig()
	imp.PageSize = "Letter"

	if err := pdfcpu_api.ImportImagesFile([]string{srcPath}, outPDF, imp, nil); err != nil {
		return fmt.Errorf("failed to import image %s to PDF: %w", flatName, err)
	}

	if err := pdfcpu_api.AddWatermarksFile(outPDF, outPDF, nil, wm, nil); err != nil {
		return fmt.Errorf("failed to add watermark to imported image PDF %s: %w", flatName, err)
	}

	return nil
}

func mergeBatch(files []string, targetDir string, counter int) error {
	fileName := fmt.Sprintf("FinalResult_Visual_Part_%d.pdf", counter)
	filePath := filepath.Join(targetDir, fileName)
	log.Printf("Flushing and writing structured batch out to file payload: %s", fileName)

	err := pdfcpu_api.MergeCreateFile(files, filePath, false, nil)
	if err != nil {
		return fmt.Errorf("failed to merge batch into %s: %w", fileName, err)
	}

	if err := cleanUpFiles(files); err != nil {
		return fmt.Errorf("failed to clean up temporary batch files: %w", err)
	}

	return nil
}
