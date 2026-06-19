package cmd

import (
	"fmt"
	"strings"

	"pdftool/internal/log"
	"pdftool/internal/util"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
)

var (
	watermarkText    string
	watermarkImage   string
	watermarkOutput  string
	watermarkPages   string
	watermarkOnTop   bool
	watermarkOpacity float64
	watermarkScale   float64
	watermarkAngle   float64
	watermarkPos     string
	watermarkFontSize int
	watermarkColor   string
)

var watermarkCmd = &cobra.Command{
	Use:   "watermark <input.pdf>",
	Short: "添加文字或图片水印",
	Long: `为 PDF 添加文字或图片水印。
支持配置位置、透明度、字号、颜色、旋转角度等。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		if watermarkText == "" && watermarkImage == "" {
			return fmt.Errorf("请指定 --text 或 --image")
		}
		if watermarkText != "" && watermarkImage != "" {
			return fmt.Errorf("不能同时指定文字和图片水印")
		}

		conf := model.NewDefaultConfiguration()

		var pageRanges []string
		var err error
		if watermarkPages != "" {
			pageRanges, err = util.ParsePageRange(watermarkPages)
			if err != nil {
				return err
			}
		}

		desc := buildWatermarkDesc()

		if watermarkText != "" {
			log.Info("添加文字水印: %s", watermarkText)
			err = api.AddTextWatermarksFile(inputFile, watermarkOutput, pageRanges, watermarkOnTop, watermarkText, desc, conf)
		} else {
			if !util.FileExists(watermarkImage) {
				return fmt.Errorf("图片文件不存在: %s", watermarkImage)
			}
			log.Info("添加图片水印: %s", watermarkImage)
			err = api.AddImageWatermarksFile(inputFile, watermarkOutput, pageRanges, watermarkOnTop, watermarkImage, desc, conf)
		}

		if err != nil {
			return fmt.Errorf("添加水印失败: %w", err)
		}

		size, _ := util.FileSize(watermarkOutput)
		log.Success("水印添加完成: %s (%s)", watermarkOutput, util.FormatSize(size))
		return nil
	},
}

func buildWatermarkDesc() string {
	parts := []string{}

	if watermarkScale != 1.0 && watermarkScale != 0 {
		parts = append(parts, fmt.Sprintf("s:%.2f", watermarkScale))
	}
	if watermarkAngle != 0 {
		parts = append(parts, fmt.Sprintf("r:%.1f", watermarkAngle))
	}
	if watermarkOpacity > 0 && watermarkOpacity < 1 {
		parts = append(parts, fmt.Sprintf("o:%.2f", watermarkOpacity))
	}
	if watermarkPos != "" {
		parts = append(parts, fmt.Sprintf("p:%s", watermarkPos))
	}
	if watermarkFontSize > 0 {
		parts = append(parts, fmt.Sprintf("fs:%d", watermarkFontSize))
	}
	if watermarkColor != "" {
		parts = append(parts, fmt.Sprintf("c:%s", watermarkColor))
	}

	return strings.Join(parts, " ")
}

var watermarkRemoveCmd = &cobra.Command{
	Use:   "remove <input.pdf>",
	Short: "移除水印",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		conf := model.NewDefaultConfiguration()

		var pageRanges []string
		if watermarkPages != "" {
			var err error
			pageRanges, err = util.ParsePageRange(watermarkPages)
			if err != nil {
				return err
			}
		}

		log.Info("移除水印...")
		err := api.RemoveWatermarksFile(inputFile, watermarkOutput, pageRanges, conf)
		if err != nil {
			return fmt.Errorf("移除水印失败: %w", err)
		}

		size, _ := util.FileSize(watermarkOutput)
		log.Success("水印移除完成: %s (%s)", watermarkOutput, util.FormatSize(size))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(watermarkCmd)
	watermarkCmd.AddCommand(watermarkRemoveCmd)

	watermarkCmd.Flags().StringVar(&watermarkText, "text", "", "文字水印内容")
	watermarkCmd.Flags().StringVar(&watermarkImage, "image", "", "图片水印文件路径")
	watermarkCmd.Flags().StringVarP(&watermarkOutput, "output", "o", "watermarked.pdf", "输出文件路径")
	watermarkCmd.Flags().StringVarP(&watermarkPages, "pages", "p", "", "指定页面范围，不指定则全部")
	watermarkCmd.Flags().BoolVar(&watermarkOnTop, "on-top", false, "水印放在内容上方（默认在下方）")
	watermarkCmd.Flags().Float64Var(&watermarkOpacity, "opacity", 0.5, "透明度 (0-1)")
	watermarkCmd.Flags().Float64Var(&watermarkScale, "scale", 1.0, "缩放比例")
	watermarkCmd.Flags().Float64Var(&watermarkAngle, "angle", 0, "旋转角度")
	watermarkCmd.Flags().StringVar(&watermarkPos, "pos", "c", "位置：tl(左上)/tr(右上)/c(居中)/bl(左下)/br(右下)")
	watermarkCmd.Flags().IntVar(&watermarkFontSize, "font-size", 24, "字体大小（文字水印）")
	watermarkCmd.Flags().StringVar(&watermarkColor, "color", "#808080", "文字颜色（十六进制）")

	watermarkRemoveCmd.Flags().StringVarP(&watermarkOutput, "output", "o", "no-watermark.pdf", "输出文件路径")
	watermarkRemoveCmd.Flags().StringVarP(&watermarkPages, "pages", "p", "", "指定页面范围")
}
