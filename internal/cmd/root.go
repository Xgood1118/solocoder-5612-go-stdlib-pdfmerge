package cmd

import (
	"pdftool/internal/log"

	"github.com/spf13/cobra"
)

var (
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "pdftool",
	Short: "PDF 工具集 - 合并、拆分、提取、旋转、水印、加密、压缩等",
	Long: `一个功能丰富的 PDF 命令行工具，支持：
  - 合并多个 PDF 文件
  - 拆分 PDF（按页数、按范围、按书签、单页）
  - 提取、删除、旋转页面
  - 添加文字/图片水印
  - 加密/解密 PDF
  - 压缩 PDF 体积
  - 管理书签和元数据
  - 查看 PDF 详细信息`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetVerbose(verbose)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "显示详细日志")
}
