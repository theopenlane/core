package gencmd

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"
	"golang.org/x/tools/imports"
)

var (
	//go:embed templates/*
	_templates embed.FS
)

const (
	templateSuffix = ".tmpl"
)

// cmd data for the template
type cmd struct {
	// Name is the name of the command
	Name string
	// ListOnly is a flag to indicate if the command is list only
	ListOnly bool
	// HistoryCmd is a flag to indicate if the command is for history
	HistoryCmd bool
}

var (
	mutationTemplates = []string{"create.tmpl", "update.tmpl", "delete.tmpl"}
)

// Generate generates the cli command files for the given command name
func Generate(cmdName string, cmdDirName string, readOnly bool, force bool) error {
	// trim any leading/trailing spaces
	cmdName = strings.Trim(cmdName, " ")

	// create new directory first
	if err := os.MkdirAll(getNewCmdDirName(cmdDirName, cmdName), os.ModePerm); err != nil {
		return err
	}

	fmt.Println("----> creating cli cmd for:", cmdName)

	templates, err := _templates.ReadDir("templates")
	if err != nil {
		return err
	}

	// generate all the cmd files
	for _, t := range templates {
		// if read only, skip the mutation templates
		if readOnly && slices.Contains(mutationTemplates, t.Name()) {
			continue
		}

		if err := generateCmdFile(cmdName, cmdDirName, t.Name(), readOnly, force); err != nil {
			return err
		}
	}

	return nil
}

// createCmd creates a new template for generating cli commands
func createCmd(name string) (*template.Template, error) {
	// function map for template
	fm := template.FuncMap{
		"ToUpperCamel": strcase.UpperCamelCase,
		"ToLowerCamel": strcase.LowerCamelCase,
		"ToLower":      strings.ToLower,
		"ToPlural":     pluralize.NewClient().Plural,
		"ToKebabCase":  strcase.KebabCase,
	}

	// create schema template
	tmpl, err := template.New(name).Funcs(fm).ParseFS(_templates, fmt.Sprintf("templates/%s", name))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// generateCmdFile generates the cmd file for the given command name and template name
func generateCmdFile(cmdName, cmdDirName, templateName string, readOnly bool, force bool) error {
	// create the template
	tmpl, err := createCmd(templateName)
	if err != nil {
		return err
	}

	// get the file name
	filePath := getFileName(cmdDirName, cmdName, templateName)

	// check if schema already exists, skip generation so we don't overwrite manual changes
	if _, err := os.Stat(filePath); err == nil && !force {
		return nil
	}

	isHistory := strings.Contains(cmdName, "History")

	// setup the data required for the template
	c := cmd{
		Name:       cmdName,
		ListOnly:   readOnly, // if read only, set the list only flag
		HistoryCmd: isHistory,
	}

	fmt.Println("----> executing template:", templateName)

	// execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		return err
	}

	out, err := imports.Process(filePath, buf.Bytes(), nil)
	if err != nil {
		return err
	}

	// create the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	if _, err := file.Write(out); err != nil {
		return err
	}

	return nil
}

// getFileName returns the file name for the cmd file to be generated
func getFileName(dir, cmdName, templateName string) string {
	// trim the trailing slash, if any
	dir = strings.TrimRight(dir, "/")

	// trim the suffix to get the file name
	fileName := strings.TrimSuffix(templateName, templateSuffix)

	fullPath := fmt.Sprintf("%s/%s/%s.go", dir, strings.ToLower(cmdName), strings.ToLower(fileName))

	return filepath.Clean(fullPath)
}

// getNewCmdDirName returns the directory name for the cmd files to be generated
func getNewCmdDirName(dir, cmdName string) string {
	// trim the trailing slash, if any
	dir = strings.TrimRight(dir, "/")

	fullPath := fmt.Sprintf("%s/%s", dir, strings.ToLower(cmdName))

	return filepath.Clean(fullPath)
}
