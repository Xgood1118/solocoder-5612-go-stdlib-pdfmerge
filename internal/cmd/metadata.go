package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"pdftool/internal/log"
	"pdftool/internal/util"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
)

var (
	metadataTitle    string
	metadataAuthor   string
	metadataSubject  string
	metadataKeywords string
	metadataOutput   string
	metadataBatch    bool
)

var metadataCmd = &cobra.Command{
	Use:   "metadata <input.pdf>",
	Short: "查看或修改 PDF 元数据",
	Long: `查看或修改 PDF 文档的元数据（标题、作者、主题、关键词）。
不带修改参数时，显示当前元数据。
使用 --batch 可批量处理多个文件。`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hasModification := metadataTitle != "" || metadataAuthor != "" || metadataSubject != "" || metadataKeywords != ""

		if metadataBatch && hasModification {
			return batchModifyMetadata(args)
		}

		if len(args) > 1 && !metadataBatch {
			return fmt.Errorf("处理多个文件请使用 --batch 参数")
		}

		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		if !hasModification {
			return showMetadata(inputFile)
		}

		return modifyMetadata(inputFile, metadataOutput)
	},
}

func showMetadata(inputFile string) error {
	conf := model.NewDefaultConfiguration()

	f, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	info, err := api.PDFInfo(f, inputFile, nil, false, conf)
	if err != nil {
		return fmt.Errorf("读取元数据失败: %w", err)
	}

	fmt.Println("=== PDF 元数据 ===")
	fmt.Printf("标题:     %s\n", info.Title)
	fmt.Printf("作者:     %s\n", info.Author)
	fmt.Printf("主题:     %s\n", info.Subject)
	fmt.Printf("关键词:   %s\n", strings.Join(info.Keywords, ", "))
	fmt.Printf("创建者:   %s\n", info.Creator)
	fmt.Printf("生成器:   %s\n", info.Producer)
	fmt.Printf("创建时间: %s\n", info.CreationDate)
	fmt.Printf("修改时间: %s\n", info.ModificationDate)

	return nil
}

func modifyMetadata(inputFile, outputFile string) error {
	if outputFile == "" {
		outputFile = inputFile
	}

	conf := model.NewDefaultConfiguration()

	props := make(map[string]string)
	if metadataTitle != "" {
		props["Title"] = metadataTitle
	}
	if metadataAuthor != "" {
		props["Author"] = metadataAuthor
	}
	if metadataSubject != "" {
		props["Subject"] = metadataSubject
	}
	if metadataKeywords != "" {
		props["Keywords"] = metadataKeywords
	}

	err := api.AddPropertiesFile(inputFile, outputFile, props, conf)
	if err != nil {
		return fmt.Errorf("修改元数据失败: %w", err)
	}

	size, _ := util.FileSize(outputFile)
	log.Success("元数据修改完成: %s (%s)", outputFile, util.FormatSize(size))
	return nil
}

func batchModifyMetadata(files []string) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	successCount := 0
	var mu sync.Mutex

	for _, file := range files {
		if !util.FileExists(file) {
			log.Warn("跳过不存在的文件: %s", file)
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(f string) {
			defer wg.Done()
			defer func() { <-sem }()

			err := modifyMetadata(f, f)
			if err != nil {
				log.Error("处理 %s 失败: %v", f, err)
				return
			}
			mu.Lock()
			successCount++
			mu.Unlock()
			log.Info("已处理: %s", f)
		}(file)
	}

	wg.Wait()
	log.Success("批量处理完成，成功 %d/%d 个文件", successCount, len(files))
	return nil
}

func init() {
	rootCmd.AddCommand(metadataCmd)
	metadataCmd.Flags().StringVar(&metadataTitle, "title", "", "设置标题")
	metadataCmd.Flags().StringVar(&metadataAuthor, "author", "", "设置作者")
	metadataCmd.Flags().StringVar(&metadataSubject, "subject", "", "设置主题")
	metadataCmd.Flags().StringVar(&metadataKeywords, "keywords", "", "设置关键词，逗号分隔")
	metadataCmd.Flags().StringVarP(&metadataOutput, "output", "o", "", "输出文件路径（仅单文件时使用）")
	metadataCmd.Flags().BoolVar(&metadataBatch, "batch", false, "批量处理多个文件（原地修改）")
}
