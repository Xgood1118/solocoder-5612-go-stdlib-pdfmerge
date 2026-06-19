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
	encryptUserPW    string
	encryptOwnerPW   string
	encryptOutput    string
	encryptNoPrint   bool
	encryptNoCopy    bool
	encryptNoModify  bool
	decryptOwnerPW   string
	decryptOutput    string
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt <input.pdf>",
	Short: "加密 PDF 文件",
	Long: `为 PDF 添加密码保护，支持设置用户密码和所有者密码，以及权限控制。
用户密码：打开 PDF 时需要输入
所有者密码：修改权限时需要输入`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		if encryptUserPW == "" && encryptOwnerPW == "" {
			return fmt.Errorf("请至少指定一个密码 --user-password 或 --owner-password")
		}

		conf := model.NewDefaultConfiguration()
		conf.UserPW = encryptUserPW
		conf.OwnerPW = encryptOwnerPW

		perms := model.PermissionsAll
		if encryptNoPrint {
			perms &^= model.PermissionPrintRev2
			perms &^= model.PermissionPrintRev3
		}
		if encryptNoCopy {
			perms &^= model.PermissionExtract
			perms &^= model.PermissionExtractRev3
		}
		if encryptNoModify {
			perms &^= model.PermissionModify
			perms &^= model.PermissionModAnnFillForm
			perms &^= model.PermissionAssembleRev3
		}
		conf.Permissions = perms

		log.Info("正在加密 PDF...")
		log.Debug("用户密码: %s", maskPassword(encryptUserPW))
		log.Debug("所有者密码: %s", maskPassword(encryptOwnerPW))

		err := api.EncryptFile(inputFile, encryptOutput, conf)
		if err != nil {
			return fmt.Errorf("加密失败: %w", err)
		}

		size, _ := util.FileSize(encryptOutput)
		log.Success("加密完成: %s (%s)", encryptOutput, util.FormatSize(size))
		return nil
	},
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt <input.pdf>",
	Short: "解密 PDF 文件",
	Long:  `使用所有者密码移除 PDF 的加密保护。`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		if !util.FileExists(inputFile) {
			return fmt.Errorf("文件不存在: %s", inputFile)
		}

		if decryptOwnerPW == "" {
			return fmt.Errorf("请使用 --owner-password 指定所有者密码")
		}

		conf := model.NewDefaultConfiguration()
		conf.OwnerPW = decryptOwnerPW

		log.Info("正在解密 PDF...")

		err := api.DecryptFile(inputFile, decryptOutput, conf)
		if err != nil {
			return fmt.Errorf("解密失败: %w", err)
		}

		size, _ := util.FileSize(decryptOutput)
		log.Success("解密完成: %s (%s)", decryptOutput, util.FormatSize(size))
		return nil
	},
}

func maskPassword(pw string) string {
	if pw == "" {
		return "(空)"
	}
	return "***"
}

func init() {
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)

	encryptCmd.Flags().StringVar(&encryptUserPW, "user-password", "", "用户密码（打开文档）")
	encryptCmd.Flags().StringVar(&encryptOwnerPW, "owner-password", "", "所有者密码（修改权限）")
	encryptCmd.Flags().StringVarP(&encryptOutput, "output", "o", "encrypted.pdf", "输出文件路径")
	encryptCmd.Flags().BoolVar(&encryptNoPrint, "no-print", false, "禁止打印")
	encryptCmd.Flags().BoolVar(&encryptNoCopy, "no-copy", false, "禁止复制内容")
	encryptCmd.Flags().BoolVar(&encryptNoModify, "no-modify", false, "禁止修改")

	decryptCmd.Flags().StringVar(&decryptOwnerPW, "owner-password", "", "所有者密码")
	decryptCmd.Flags().StringVarP(&decryptOutput, "output", "o", "decrypted.pdf", "输出文件路径")
}
