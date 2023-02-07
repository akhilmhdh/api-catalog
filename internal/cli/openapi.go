package cli

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/invopop/yaml"
)

func ValidateOpenAPI(raw []byte, apiSchemaFile map[string]any, logger *CliLogger) error {
	if val, ok := apiSchemaFile["swagger"]; ok && (strings.HasPrefix(val.(string), "2.") || val.(string) == "2") {
		logger.Info("Detected OpenAPI v2 schema")
		logger.Info("Converting the schema to v3")
		// convert to OpenAPIv3
		var docv2 openapi2.T
		if err := yaml.Unmarshal(raw, &docv2); err != nil {
			return err
		}
		docv3, err := openapi2conv.ToV3(&docv2)
		if err != nil {
			return nil
		}

		logger.Success("Converted v2 to v3 schema")
		loader := openapi3.NewLoader()

		logger.Info("Validating by OpenAPI schema specs")
		if err = docv3.Validate(loader.Context); err != nil {
			logger.Error("Failed to meet OpenAPI spec")
			fmt.Println(err)
		}

		// now we need the openapi v3 version of map[strings]
		if raw, err = docv3.MarshalJSON(); err != nil {
			return err
		}

		if err = yaml.Unmarshal(raw, &apiSchemaFile); err != nil {
			return err
		}

		return nil
	}
	logger.Info("Detected OpenAPI v3 schema")
	loader := openapi3.NewLoader()
	docv3, err := loader.LoadFromData(raw)
	if err != nil {
		return err
	}

	logger.Info("Validating by OpenAPI schema specs")
	if err := docv3.Validate(loader.Context); err != nil {
		logger.Error("Failed to meet OpenAPI spec")
		fmt.Println(err)

	}

	return nil
}
