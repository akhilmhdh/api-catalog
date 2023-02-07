import { isCasing } from "apic/strings";

export default function (config, options = {}) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;
  const reqBodyCasing = options?.req_body_casing || "camelcase";
  const paramsCasing = options?.params_casing || "camelcase";

  Object.keys(config.schema.paths).forEach((path) => {
    Object.keys(config.schema.paths[path]).forEach((method) => {
      (config.schema.paths[path][method].parameters || []).forEach((param) => {
        numberOfResponses++;
        if (!isCasing(paramsCasing, param.name)) {
          numbnerOfFalseResponses++;
          config.report({
            message: `Invalid casing for ${param.name} of ${param.in}`,
            path: path,
            method: method,
          });
        }
      });
    });
  });

  Object.keys(config.schema.components.schemas).forEach((schema) => {
    Object.keys(config.schema.components.schemas[schema].properties).forEach(
      (property) => {
        numberOfResponses++;
        if (!isCasing(reqBodyCasing, property)) {
          numbnerOfFalseResponses++;
          config.report({
            message: `Invalid casing for ${property} of schema ${schema}`,
            path: "Nil",
            method: "Nil",
          });
        }
      }
    );
  });

  // if number goes to negative
  const score =
    (Math.max(numberOfResponses - numbnerOfFalseResponses, 0) /
      numberOfResponses) *
    100;
  config.setScore("quality", score);
}
