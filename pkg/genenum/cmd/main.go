//go:build genenum

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/stoewer/go-strcase"
)

var (
	enumName   string
	enumValues string

	errRequiredValues = errors.New("enum name and values are required")
)

func promptIfEmpty() {
	reader := bufio.NewReader(os.Stdin)

	if enumName == "" {
		fmt.Print("Enter enum name (e.g. TaskStatus): ")

		nameInput, _ := reader.ReadString('\n')
		enumName = strings.TrimSpace(nameInput)
	}

	if enumValues == "" {
		fmt.Print("Enter enum values (comma-separated, e.g. OPEN,IN_PROGRESS): ")

		valuesInput, _ := reader.ReadString('\n')
		enumValues = strings.TrimSpace(valuesInput)
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "enumgen",
		Short: "Enum code generator for Go",
		RunE: func(_ *cobra.Command, _ []string) error {
			promptIfEmpty()

			if enumName == "" || enumValues == "" {
				return errRequiredValues
			}

			return generateEnum(enumName, strings.Split(enumValues, ","))
		},
	}

	rootCmd.Flags().StringVar(&enumName, "name", "", "Name of the enum type (e.g. TaskStatus)")
	rootCmd.Flags().StringVar(&enumValues, "values", "", "Comma-separated list of enum values (e.g. OPEN,IN_PROGRESS)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("enum created")
}

func generateEnum(name string, values []string) error {
	funcMap := template.FuncMap{
		"ToCamel":         strcase.UpperCamelCase,
		"ToUpper":         strings.ToUpper,
		"lowerToSentence": lowerToSentence,
	}

	tmplBytes, err := os.ReadFile("../pkg/genenum/cmd/templates/enums.tmpl")
	if err != nil {
		return err
	}

	tmpl, err := template.New("enum").Funcs(funcMap).Parse(string(tmplBytes))
	if err != nil {
		return err
	}

	outputFile := strcase.SnakeCase(strings.ToLower(name)) + ".go"

	file, err := os.Create("../common/enums/" + outputFile)
	if err != nil {
		return err
	}

	defer file.Close()

	seen := map[string]struct{}{}
	uniqueValues := []string{}

	for _, v := range values {
		val := strings.ToUpper(v)
		if _, exists := seen[val]; !exists {
			seen[val] = struct{}{}

			uniqueValues = append(uniqueValues, val)
		}
	}

	data := struct {
		Name   string
		Values []string
	}{
		Name:   name,
		Values: uniqueValues,
	}

	return tmpl.Execute(file, data)
}

func lowerToSentence(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ToLower(s)

	return s
}
