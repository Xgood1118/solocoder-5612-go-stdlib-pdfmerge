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
	compressOutput string
)

var compressCmd = &cobra.Command{
	Use:   "compress <input.pdf>",
	Short: "压缩 PDF 文件",
	Long: `压缩 PDF 文件，减小文件体积。
包括优化图片、清理冗余对象等。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		origSize, _ := util.FileSize(inputFile)
		log.Info("正在压缩 PDF...")
		log.Debug("原始大小: %s", util.FormatSize(origSize))

		conf := model.NewDefaultConfiguration()

		err := api.OptimizeFile(inputFile, compressOutput, conf)
		if err != nil {
			return fmt.Errorf("压缩失败: %w", err)
		}

		newSize, _ := util.FileSize(compressOutput)
		saved := origSize - newSize
		savedPct := float64(0)
		if origSize > 0 {
			savedPct = float64(saved) / float64(origSize) * 100
		}

		log.Success("压缩完成: %s", compressOutput)
		log.Info("  原始大小: %s", util.FormatSize(origSize))
		log.Info("  压缩后:   %s", util.FormatSize(newSize))
		log.Info("  节省:     %s (%.1f%%)", util.FormatSize(saved), savedPct)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(compressCmd)
	compressCmd.Flags().StringVarP(&compressOutput, "output", "o", "compressed.pdf", "输出文件路径")
}
