package template

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"gitlab-cli-sdk/pkg/types"
)

// RenderTemplate 使用模板文件渲染输出
func RenderTemplate(templateFile string, data *types.OutputConfig) (string, error) {
	// 读取模板文件
	templateContent, err := os.ReadFile(templateFile)
	if err != nil {
		return "", fmt.Errorf("read template file: %w", err)
	}

	// 创建模板
	tmpl, err := template.New("output").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// SaveTemplateOutput 使用模板渲染并保存到文件
func SaveTemplateOutput(templateFile, outputFile string, data *types.OutputConfig) error {
	// 渲染模板
	result, err := RenderTemplate(templateFile, data)
	if err != nil {
		return err
	}

	// 保存到文件
	if err := os.WriteFile(outputFile, []byte(result), 0644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}

	return nil
}
