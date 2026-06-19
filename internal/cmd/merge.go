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
	mergeOutput string
)

var mergeCmd = &cobra.Command{
	Use:   "merge [pdf1 pdf2 ...",
	Short: "合并多个 PDF 文件",
	Long:  `按顺序合并多个 PDF 文件为一个 PDF 文件。`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, f := range args {
			if !util.FileExists(f) {
				return fmt.Errorf("文件不存在: %s", f)
			}
		}

		conf := model.NewDefaultConfiguration()

		log.Info("正在合并 %d 个 PDF 文件...", len(args))
		for i, f := range args {
			log.Debug("  %d. %s", i+1, f)
		}

		err := api.MergeCreateFile(args, mergeOutput, false, conf)
		if err != nil {
			return fmt.Errorf("合并失败: %w", err)
		}

		size, _ := util.FileSize(mergeOutput)
		log.Success("合并完成: %s (%s)", mergeOutput, util.FormatSize(size))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringVarP(&mergeOutput, "output", "o", "merged.pdf", "输出文件路径")
}
