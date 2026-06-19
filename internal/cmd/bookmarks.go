package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"pdftool/internal/log"
	"pdftool/internal/util"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
)

var (
	bookmarksImport string
	bookmarksExport string
)

var bookmarksCmd = &cobra.Command{
	Use:   "bookmarks <input.pdf>",
	Short: "管理 PDF 书签",
	Long: `列出、导入或导出 PDF 书签。
不带参数时列出书签。
--import 从 CSV 文件导入书签。
--export 导出书签到 JSON 文件。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		if bookmarksImport != "" {
			return importBookmarks(inputFile)
		}

		if bookmarksExport != "" {
			return exportBookmarks(inputFile)
		}

		return listBookmarks(inputFile)
	},
}

func listBookmarks(inputFile string) error {
	conf := model.NewDefaultConfiguration()

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
		fmt.Println("PDF 没有书签")
		return nil
	}

	fmt.Println("=== PDF 书签 ===")
	printBookmarks(bms, 0)
	fmt.Printf("\n共 %d 个书签\n", len(bms))

	return nil
}

func printBookmarks(bms []pdfcpu.Bookmark, level int) {
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}
	for _, bm := range bms {
		fmt.Printf("%s- %s (页 %d)\n", indent, bm.Title, bm.PageFrom)
		if len(bm.Kids) > 0 {
			printBookmarks(bm.Kids, level+1)
		}
	}
}

type bookmarkJSON struct {
	Title string         `json:"title"`
	Page  int            `json:"page"`
	Kids  []bookmarkJSON `json:"kids,omitempty"`
}

func importBookmarks(inputFile string) error {
	if !util.FileExists(bookmarksImport) {
		return fmt.Errorf("书签文件不存在: %s", bookmarksImport)
	}

	f, err := os.Open(bookmarksImport)
	if err != nil {
		return fmt.Errorf("打开书签文件失败: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取 CSV 失败: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV 文件为空")
	}

	jsonData, err := buildBookmarksJSON(records)
	if err != nil {
		return fmt.Errorf("生成书签 JSON 失败: %w", err)
	}

	conf := model.NewDefaultConfiguration()
	outputFile := "bookmarks_imported.pdf"

	log.Info("正在导入书签...")
	log.Debug("书签数量: %d", len(records))
	log.Debug("JSON: %s", string(jsonData))

	tmpFile, err := os.CreateTemp("", "bookmarks_*.json")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := tmpFile.Write(jsonData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("写入临时文件失败: %w", err)
	}
	tmpFile.Close()

	err = api.ImportBookmarksFile(inputFile, tmpName, outputFile, true, conf)
	if err != nil {
		return fmt.Errorf("导入书签失败: %w", err)
	}

	size, _ := util.FileSize(outputFile)
	log.Success("书签导入完成: %s (%s)", outputFile, util.FormatSize(size))
	return nil
}

func buildBookmarksJSON(records [][]string) ([]byte, error) {
	root := []bookmarkJSON{}
	var lastAtLevel []*bookmarkJSON

	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		title := record[0]
		page, err := strconv.Atoi(record[1])
		if err != nil || page < 1 {
			log.Warn("跳过无效的页码: %s", record[1])
			continue
		}
		level := 1
		if len(record) >= 3 {
			level, _ = strconv.Atoi(record[2])
		}
		if level < 1 {
			level = 1
		}

		node := bookmarkJSON{Title: title, Page: page}

		for len(lastAtLevel) >= level {
			lastAtLevel = lastAtLevel[:len(lastAtLevel)-1]
		}

		if level == 1 {
			root = append(root, node)
			if len(lastAtLevel) == 0 {
				lastAtLevel = append(lastAtLevel, &root[len(root)-1])
			} else {
				lastAtLevel[0] = &root[len(root)-1]
			}
		} else {
			parent := lastAtLevel[level-2]
			parent.Kids = append(parent.Kids, node)
			if len(lastAtLevel) >= level {
				lastAtLevel[level-1] = &parent.Kids[len(parent.Kids)-1]
			} else {
				lastAtLevel = append(lastAtLevel, &parent.Kids[len(parent.Kids)-1])
			}
		}
	}

	if len(root) == 0 {
		return nil, fmt.Errorf("没有有效的书签数据")
	}

	return json.MarshalIndent(root, "", "  ")
}

func exportBookmarks(inputFile string) error {
	conf := model.NewDefaultConfiguration()

	log.Info("正在导出书签到: %s", bookmarksExport)

	err := api.ExportBookmarksFile(inputFile, bookmarksExport, conf)
	if err != nil {
		return fmt.Errorf("导出书签失败: %w", err)
	}

	log.Success("书签导出完成: %s", bookmarksExport)
	return nil
}

func init() {
	rootCmd.AddCommand(bookmarksCmd)
	bookmarksCmd.Flags().StringVar(&bookmarksImport, "import", "", "从 CSV 文件导入书签 (格式: 标题,页码,层级)")
	bookmarksCmd.Flags().StringVar(&bookmarksExport, "export", "", "导出书签到 JSON 文件")
}
