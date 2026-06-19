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
	deletePages  string
	deleteOutput string
)

var deleteCmd = &cobra.Command{
	Use:   "delete <input.pdf>",
	Short: "删除指定页面",
	Long: `从 PDF 中删除指定页面，输出不包含这些页面的新 PDF。
页面范围语法："1,3,5-10,15-20"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		if deletePages == "" {
			return fmt.Errorf("请使用 --pages 指定要删除的页面范围")
		}

		pageRanges, err := util.ParsePageRange(deletePages)
		if err != nil {
			return err
		}

		conf := model.NewDefaultConfiguration()

		log.Info("正在删除页面: %s", deletePages)
		log.Debug("输入文件: %s", inputFile)
		log.Debug("输出文件: %s", deleteOutput)

		err = api.RemovePagesFile(inputFile, deleteOutput, pageRanges, conf)
		if err != nil {
			return fmt.Errorf("删除页面失败: %w", err)
		}

		size, _ := util.FileSize(deleteOutput)
		log.Success("页面删除完成: %s (%s)", deleteOutput, util.FormatSize(size))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&deletePages, "pages", "p", "", "要删除的页面范围，如 \"1,3,5-10\"")
	deleteCmd.Flags().StringVarP(&deleteOutput, "output", "o", "deleted.pdf", "输出文件路径")
}
