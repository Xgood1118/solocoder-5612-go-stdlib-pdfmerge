package cmd

import (
	"fmt"
	"os"
	"strings"

	"pdftool/internal/util"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
)

var (
	infoShowFonts bool
)

var infoCmd = &cobra.Command{
	Use:   "info <input.pdf>",
	Short: "显示 PDF 文件信息",
	Long: `显示 PDF 文件的详细信息，包括：
  - 页数、文件大小、PDF 版本
  - 是否加密、是否带表单、是否带附件
  - 标题、作者、主题、关键词等元数据
  - 字体列表`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		conf := model.NewDefaultConfiguration()

		f, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("打开文件失败: %w", err)
		}
		defer f.Close()

		info, err := api.PDFInfo(f, inputFile, nil, infoShowFonts, conf)
		if err != nil {
			return fmt.Errorf("获取 PDF 信息失败: %w", err)
		}

		fileSize, _ := util.FileSize(inputFile)

		fmt.Println("=== PDF 信息 ===")
		fmt.Printf("文件名:          %s\n", info.FileName)
		fmt.Printf("文件大小:        %s\n", util.FormatSize(fileSize))
		fmt.Printf("PDF 版本:        %s\n", info.Version)
		fmt.Printf("页数:            %d\n", info.PageCount)
		fmt.Println()

		fmt.Println("=== 文档属性 ===")
		fmt.Printf("标题:            %s\n", info.Title)
		fmt.Printf("作者:            %s\n", info.Author)
		fmt.Printf("主题:            %s\n", info.Subject)
		fmt.Printf("创建者:          %s\n", info.Creator)
		fmt.Printf("生成器:          %s\n", info.Producer)
		fmt.Printf("创建时间:        %s\n", info.CreationDate)
		fmt.Printf("修改时间:        %s\n", info.ModificationDate)
		if len(info.Keywords) > 0 {
			fmt.Printf("关键词:          %s\n", strings.Join(info.Keywords, ", "))
		}
		fmt.Println()

		fmt.Println("=== 文档特征 ===")
		fmt.Printf("已加密:          %v\n", info.Encrypted)
		fmt.Printf("带表单:          %v\n", info.Form)
		fmt.Printf("带书签:          %v\n", info.Outlines)
		fmt.Printf("带附件:          %v\n", len(info.Attachments) > 0)
		fmt.Printf("带水印:          %v\n", info.Watermarked)
		fmt.Printf("带缩略图:        %v\n", info.Thumbnails)
		fmt.Printf("线性化:          %v\n", info.Linearized)
		fmt.Printf("标签化:          %v\n", info.Tagged)
		fmt.Printf("有签名:          %v\n", info.Signatures)
		fmt.Println()

		if infoShowFonts && len(info.Fonts) > 0 {
			fmt.Println("=== 字体列表 ===")
			for i, font := range info.Fonts {
				fmt.Printf("  %d. 名称: %s, 类型: %s, 嵌入: %v\n",
					i+1, font.Name, font.Type, font.Embedded)
			}
			fmt.Println()
		}

		if info.Encrypted {
			fmt.Println("=== 权限 ===")
			fmt.Printf("权限值:          %d\n", info.Permissions)
			printPermissions(info.Permissions)
		}

		return nil
	},
}

func printPermissions(perms int) {
	flags := []struct {
		bit   int
		label string
	}{
		{3, "打印"},
		{4, "修改内容"},
		{5, "复制内容"},
		{6, "添加/修改注释"},
		{9, "填写表单"},
		{10, "提取内容（无障碍）"},
		{11, "组装文档"},
		{12, "高分辨率打印"},
	}
	for _, f := range flags {
		allowed := perms&(1<<(f.bit-1)) != 0
		fmt.Printf("  %s: %v\n", f.label, allowed)
	}
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().BoolVar(&infoShowFonts, "fonts", false, "显示字体列表")
}
