package cli

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/invopop/yaml"
)

func OpenAPIValidator(raw []byte, apiSchemaFile map[string]any) error {
	if val, ok := apiSchemaFile["swagger"]; ok && (strings.HasPrefix(val.(string), "2.") || val.(string) == "2") {
		// convert to OpenAPIv3
		var docv2 openapi2.T
		if err := yaml.Unmarshal(raw, &docv2); err != nil {
			return err
		}
		docv3, err := openapi2conv.ToV3(&docv2)
		if err != nil {
			return nil
		}

		// loader := openapi3.NewLoader()
		// if err = docv3.Validate(loader.Context); err != nil {
		// 	return err
		// }

		// now we need the openapi v3 version of map[strings]
		if raw, err = docv3.MarshalJSON(); err != nil {
			return err
		}

		if err = yaml.Unmarshal(raw, &apiSchemaFile); err != nil {
			return err
		}

		return nil
	}

	loader := openapi3.NewLoader()
	docv3, err := loader.LoadFromData(raw)
	if err != nil {
		return err
	}

	if err := docv3.Validate(loader.Context); err != nil {
		return err
	}

	return nil

}
