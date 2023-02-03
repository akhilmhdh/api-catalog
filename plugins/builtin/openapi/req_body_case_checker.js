const snakeCaseRegex = /^[a-z0-9]+(?:_[a-z0-9]+)*$/;
const camelCaseRegex = /^[a-z]+(?:[A-Z0-9]+[a-z0-9]+[A-Za-z0-9]*)*$/;
const pascalCaseRegex = /^(?:[A-Z][a-z0-9]+)(?:[A-Z]+[a-z0-9]*)*$/;
const kebabCaseRegex = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

function isCamelCase(word) {
  return camelCaseRegex.test(word);
}

function isPascalCase(word) {
  return pascalCaseRegex.test(word);
}

function isSnakeCase(word) {
  return snakeCaseRegex.test(word);
}

function isKebabCase(word) {
  return kebabCaseRegex.test(word);
}

function getCaseCheckerFn(type) {
  switch (type) {
    case "camelcase":
      return isCamelCase;
    case "snakecase":
      return isSnakeCase;
    case "pascalcase":
      return isPascalCase;
    case "kebabcase":
      return isKebabCase;
    default:
      return isCamelCase;
  }
}

export default function (config, options = {}) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;
  const checkerFn = getCaseCheckerFn(options?.casing);

  Object.keys(config.schema.paths).forEach((path) => {
    Object.keys(config.schema.paths[path]).forEach((method) => {
      (config.schema.paths[path][method].parameters || []).forEach((param) => {
        numberOfResponses++;
        if (!checkerFn(param.name)) {
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
        if (!checkerFn(property)) {
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
