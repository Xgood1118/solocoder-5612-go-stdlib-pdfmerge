package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"pdftool/internal/log"
	"pdftool/internal/util"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
)

var (
	watermarkText     string
	watermarkImage    string
	watermarkOutput   string
	watermarkPages    string
	watermarkOnTop    bool
	watermarkOpacity  float64
	watermarkScale    float64
	watermarkAngle    float64
	watermarkPos      string
	watermarkFontSize int
	watermarkColor    string
	watermarkFontName string
	watermarkDiagonal int
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
		log.Debug("水印配置: %s", desc)

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

func parseHexColor(hex string) (float64, float64, float64, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %s", hex)
	}

	r, err := strconv.ParseInt(hex[0:2], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}

	return float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0, nil
}

func buildWatermarkDesc() string {
	parts := []string{}

	parts = append(parts, fmt.Sprintf("f:%s", watermarkFontName))
	parts = append(parts, fmt.Sprintf("p:%d", watermarkFontSize))
	parts = append(parts, fmt.Sprintf("s:%.2f rel", watermarkScale))

	if watermarkColor != "" {
		r, g, b, err := parseHexColor(watermarkColor)
		if err == nil {
			parts = append(parts, fmt.Sprintf("c:%.2f %.2f %.2f", r, g, b))
		} else {
			parts = append(parts, "c:0.50 0.50 0.50")
		}
	}

	if watermarkAngle != 0 {
		parts = append(parts, fmt.Sprintf("r:%.1f", watermarkAngle))
	} else if watermarkDiagonal > 0 {
		parts = append(parts, fmt.Sprintf("d:%d", watermarkDiagonal))
	}

	parts = append(parts, fmt.Sprintf("o:%.2f", watermarkOpacity))
	parts = append(parts, "m:0")

	if watermarkPos != "" {
		parts = append(parts, fmt.Sprintf("po:%s", watermarkPos))
	}

	return strings.Join(parts, ", ")
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
	watermarkCmd.Flags().Float64Var(&watermarkScale, "scale", 0.5, "缩放比例 (相对页面)")
	watermarkCmd.Flags().Float64Var(&watermarkAngle, "angle", 0, "旋转角度 (-180 到 180)")
	watermarkCmd.Flags().IntVar(&watermarkDiagonal, "diagonal", 0, "对角线 (1=左下到右上, 2=左上到右下)")
	watermarkCmd.Flags().StringVar(&watermarkPos, "pos", "", "位置：tl/tc/tr/l/c/r/bl/bc/br")
	watermarkCmd.Flags().IntVar(&watermarkFontSize, "font-size", 24, "字体大小（文字水印）")
	watermarkCmd.Flags().StringVar(&watermarkFontName, "font-name", "Helvetica", "字体名称：Helvetica/Times-Roman/Courier")
	watermarkCmd.Flags().StringVar(&watermarkColor, "color", "#808080", "文字颜色（十六进制）")

	watermarkRemoveCmd.Flags().StringVarP(&watermarkOutput, "output", "o", "no-watermark.pdf", "输出文件路径")
	watermarkRemoveCmd.Flags().StringVarP(&watermarkPages, "pages", "p", "", "指定页面范围")
}
