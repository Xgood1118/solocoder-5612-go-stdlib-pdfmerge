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
	rotateAngle  int
	rotatePages  string
	rotateOutput string
)

var rotateCmd = &cobra.Command{
	Use:   "rotate <input.pdf>",
	Short: "旋转指定页面",
	Long: `旋转 PDF 中的指定页面，支持 90/180/270 度旋转。
页面范围语法："1,3,5-10,15-20"，不指定则旋转所有页面`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		if rotateAngle%90 != 0 || rotateAngle < 0 || rotateAngle > 270 {
			return fmt.Errorf("旋转角度必须是 0/90/180/270 度")
		}

		var pageRanges []string
		var err error
		if rotatePages != "" {
			pageRanges, err = util.ParsePageRange(rotatePages)
			if err != nil {
				return err
			}
		}

		conf := model.NewDefaultConfiguration()

		log.Info("正在旋转页面 %d 度", rotateAngle)
		if rotatePages != "" {
			log.Debug("页面范围: %s", rotatePages)
		} else {
			log.Debug("页面范围: 全部")
		}
		log.Debug("输出文件: %s", rotateOutput)

		err = api.RotateFile(inputFile, rotateOutput, rotateAngle, pageRanges, conf)
		if err != nil {
			return fmt.Errorf("旋转页面失败: %w", err)
		}

		size, _ := util.FileSize(rotateOutput)
		log.Success("页面旋转完成: %s (%s)", rotateOutput, util.FormatSize(size))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rotateCmd)
	rotateCmd.Flags().IntVarP(&rotateAngle, "angle", "a", 90, "旋转角度：90/180/270")
	rotateCmd.Flags().StringVarP(&rotatePages, "pages", "p", "", "要旋转的页面范围，如 \"1,3,5-10\"，不指定则全部")
	rotateCmd.Flags().StringVarP(&rotateOutput, "output", "o", "rotated.pdf", "输出文件路径")
}
