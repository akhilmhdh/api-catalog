package cli

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Plugin API Design Spec Sheet
// Unique Name
// tags - string[]
// scores - Array(Object({tag:string, value: string}}))
// metadata - Array(Object({ key:string, value:string, type:string }))

//go:embed babel.min.js
var babelBundle string

type ApiCatalogConfig struct {
	Title string
}

func Run() {
	// cli flags
	var apiType string
	var apiURL string
	var configFilePath string

	rootCmd := &cobra.Command{
		Use:   "apic",
		Short: "One shot cli for your api schema security,performance and quality check",
		Run: func(cmd *cobra.Command, args []string) {
			// find config file and load up the config
			var config ApiCatalogConfig

			viper.AddConfigPath(configFilePath)
			viper.SetConfigName("apic")

			if err := viper.ReadInConfig(); err != nil {
				log.Fatal("File not found", err)
			}

			if err := viper.Unmarshal(&config); err != nil {
				log.Fatal("Error in config file", err)
			}

			fmt.Println(config.Title)
			fmt.Println(apiType)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&apiType, "apiType", "a", "", "Your API Type. Allowed values: rest | graphql")
	rootCmd.MarkPersistentFlagRequired("apiType")

	rootCmd.PersistentFlags().StringVar(&apiURL, "url", "", "URL or local file containing spec sheet")
	rootCmd.MarkPersistentFlagRequired("url")

	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", ".", "Path to apic configuration file")

	reqistry := new(require.Registry)

	runtime := goja.New()
	reqistry.Enable(runtime)
	console.Enable(runtime)

	babelProgram, err := goja.Compile("babel.js", babelBundle, false)

	if err != nil {
		panic(err)
	}

	_, err = runtime.RunProgram(babelProgram)
	if err != nil {
		panic(err)
	}

	var babelTransformer goja.Callable

	babel := runtime.Get("Babel")
	if err := runtime.ExportTo(babel.ToObject(runtime).Get("transform"), &babelTransformer); err != nil {
		panic(err)
	}

	jsRawCode := `
   		const b = 10;
    `

	v, err := babelTransformer(babel, runtime.ToValue(jsRawCode), runtime.ToValue(map[string]interface{}{
		"presets": []string{"env"},
	}))

	if err != nil {
		panic(err)
	}

	code := v.ToObject(runtime).Get("code").String()
	if _, err := runtime.RunString(code); err != nil {
		panic(err)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
