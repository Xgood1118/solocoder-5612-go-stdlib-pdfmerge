package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pdftool/internal/log"
	"pdftool/internal/util"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
)

var (
	splitPagesPerFile int
	splitPages        string
	splitSingle       bool
	splitByBookmarks  bool
	splitOutputDir    string
	splitPrefix       string
)

var splitCmd = &cobra.Command{
	Use:   "split <input.pdf>",
	Short: "拆分 PDF 文件",
	Long: `将一个 PDF 文件拆分为多个文件，支持多种拆分方式：
  - 按页数拆分 (--pages-per-file N)
  - 按页面范围拆分 (--pages "1-3,5,7-9")
  - 拆分为单页 (--single)
  - 按书签拆分 (--bookmarks)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		conf := model.NewDefaultConfiguration()

		if splitOutputDir == "" {
			splitOutputDir = "."
		}
		if err := util.EnsureDir(splitOutputDir); err != nil {
			return fmt.Errorf("创建输出目录失败: %w", err)
		}

		modeCount := 0
		if splitPagesPerFile > 0 {
			modeCount++
		}
		if splitPages != "" {
			modeCount++
		}
		if splitSingle {
			modeCount++
		}
		if splitByBookmarks {
			modeCount++
		}

		if modeCount == 0 {
			splitSingle = true
		} else if modeCount > 1 {
			return fmt.Errorf("只能指定一种拆分模式")
		}

		switch {
		case splitPagesPerFile > 0:
			return splitBySpan(inputFile, splitPagesPerFile, conf)
		case splitPages != "":
			return splitByPageRange(inputFile, conf)
		case splitSingle:
			return splitBySpan(inputFile, 1, conf)
		case splitByBookmarks:
			return splitByBookmark(inputFile, conf)
		}
		return nil
	},
}

func splitBySpan(inputFile string, span int, conf *model.Configuration) error {
	log.Info("按每 %d 页拆分: %s", span, inputFile)

	err := api.SplitFile(inputFile, splitOutputDir, span, conf)
	if err != nil {
		return fmt.Errorf("拆分失败: %w", err)
	}

	files, err := listSplitFiles(inputFile)
	if err == nil {
		log.Success("拆分完成，生成 %d 个文件", len(files))
		for _, f := range files {
			size, _ := util.FileSize(filepath.Join(splitOutputDir, f))
			log.Debug("  %s (%s)", f, util.FormatSize(size))
		}
	} else {
		log.Success("拆分完成")
	}
	return nil
}

func splitByPageRange(inputFile string, conf *model.Configuration) error {
	log.Info("按页面范围拆分: %s", inputFile)
	log.Debug("页面范围: %s", splitPages)

	pageRanges, err := util.ParsePageRange(splitPages)
	if err != nil {
		return err
	}

	base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))

	for i, pr := range pageRanges {
		outFile := filepath.Join(splitOutputDir, fmt.Sprintf("%s_%d.pdf", base, i+1))
		log.Debug("提取范围 %s -> %s", pr, outFile)

		err := api.CollectFile(inputFile, outFile, []string{pr}, conf)
		if err != nil {
			return fmt.Errorf("提取范围 %s 失败: %w", pr, err)
		}

		size, _ := util.FileSize(outFile)
		log.Debug("  完成: %s (%s)", outFile, util.FormatSize(size))
	}

	log.Success("按页面范围拆分完成，生成 %d 个文件", len(pageRanges))
	return nil
}

func splitByBookmark(inputFile string, conf *model.Configuration) error {
	log.Info("按书签层级拆分: %s", inputFile)

	f, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	bms, err := api.Bookmarks(f, conf)
	if err != nil {
		return fmt.Errorf("读取书签失败: %w", err)
	}

	if len(bms) == 0 {
		return fmt.Errorf("PDF 没有书签")
	}

	pageCount, err := api.PageCount(f, conf)
	if err != nil {
		return fmt.Errorf("获取页数失败: %w", err)
	}

	allBookmarks := flattenBookmarks(bms)
	if len(allBookmarks) == 0 {
		return fmt.Errorf("没有可用于拆分的书签")
	}

	base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	safeChars := func(s string) string {
		s = strings.ReplaceAll(s, "/", "_")
		s = strings.ReplaceAll(s, "\\", "_")
		s = strings.ReplaceAll(s, ":", "_")
		s = strings.ReplaceAll(s, "*", "_")
		s = strings.ReplaceAll(s, "?", "_")
		s = strings.ReplaceAll(s, "\"", "_")
		s = strings.ReplaceAll(s, "<", "_")
		s = strings.ReplaceAll(s, ">", "_")
		s = strings.ReplaceAll(s, "|", "_")
		return s
	}

	for i, bm := range allBookmarks {
		startPage := bm.PageFrom
		endPage := pageCount
		if i < len(allBookmarks)-1 {
			endPage = allBookmarks[i+1].PageFrom - 1
		}
		if endPage < startPage {
			endPage = startPage
		}

		pageRange := fmt.Sprintf("%d-%d", startPage, endPage)
		safeTitle := safeChars(bm.Title)
		if safeTitle == "" {
			safeTitle = fmt.Sprintf("bookmark_%d", i+1)
		}
		outFile := filepath.Join(splitOutputDir, fmt.Sprintf("%s_%02d_%s.pdf", base, i+1, safeTitle))

		log.Debug("提取书签 \"%s\" 页 %s -> %s", bm.Title, pageRange, outFile)

		err := api.CollectFile(inputFile, outFile, []string{pageRange}, conf)
		if err != nil {
			log.Warn("提取书签 \"%s\" 失败: %v", bm.Title, err)
			continue
		}

		size, _ := util.FileSize(outFile)
		log.Debug("  完成: %s (%s)", outFile, util.FormatSize(size))
	}

	log.Success("按书签拆分完成，生成 %d 个文件", len(allBookmarks))
	return nil
}

func flattenBookmarks(bms []pdfcpu.Bookmark) []pdfcpu.Bookmark {
	var result []pdfcpu.Bookmark
	for _, bm := range bms {
		result = append(result, bm)
		if len(bm.Kids) > 0 {
			result = append(result, flattenBookmarks(bm.Kids)...)
		}
	}
	return result
}

func listSplitFiles(inputFile string) ([]string, error) {
	base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	var files []string
	entries, err := os.ReadDir(splitOutputDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), base+"_") && strings.HasSuffix(e.Name(), ".pdf") {
			files = append(files, e.Name())
		}
	}
	return files, nil
}

func init() {
	rootCmd.AddCommand(splitCmd)
	splitCmd.Flags().IntVar(&splitPagesPerFile, "pages-per-file", 0, "每 N 页拆分为一个文件")
	splitCmd.Flags().StringVar(&splitPages, "pages", "", "按页面范围拆分，如 \"1-3,5,7-9\"")
	splitCmd.Flags().BoolVar(&splitSingle, "single", false, "拆分为单页 PDF")
	splitCmd.Flags().BoolVar(&splitByBookmarks, "bookmarks", false, "按书签层级拆分")
	splitCmd.Flags().StringVar(&splitOutputDir, "output-dir", ".", "输出目录")
}
