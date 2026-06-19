package cmd

import (
	"fmt"

	"pdftool/internal/log"
	"pdftool/internal/util"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
)

var (
	extractPages  string
	extractOutput string
)

var extractCmd = &cobra.Command{
	Use:   "extract <input.pdf>",
	Short: "提取指定页面到新 PDF",
	Long: `从 PDF 中提取指定页面到新的 PDF 文件。
页面范围语法："1,3,5-10,15-20"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		if extractPages == "" {
			return fmt.Errorf("请使用 --pages 指定要提取的页面范围")
		}

		pageRanges, err := util.ParsePageRange(extractPages)
		if err != nil {
			return err
		}

		conf := model.NewDefaultConfiguration()

		log.Info("正在提取页面: %s", extractPages)
		log.Debug("输入文件: %s", inputFile)
		log.Debug("输出文件: %s", extractOutput)

		err = api.CollectFile(inputFile, extractOutput, pageRanges, conf)
		if err != nil {
			return fmt.Errorf("提取页面失败: %w", err)
		}

		size, _ := util.FileSize(extractOutput)
		log.Success("页面提取完成: %s (%s)", extractOutput, util.FormatSize(size))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().StringVarP(&extractPages, "pages", "p", "", "要提取的页面范围，如 \"1,3,5-10\"")
	extractCmd.Flags().StringVarP(&extractOutput, "output", "o", "extracted.pdf", "输出文件路径")
}
