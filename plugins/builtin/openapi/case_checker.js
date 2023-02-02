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

// for dynamic parameters like /pets/{something}
function removeParenthesis(path) {
  if (path[0] === "{" && path[path.length - 1] === "}") {
    return path.slice(1, -1);
  }

  return path;
}

export default function (config, options) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;

  const checkerFn = getCaseCheckerFn(options.casing);

  try {
    Object.keys(config.schema.paths).forEach((path) => {
      path
        .split("/")
        .filter(Boolean)
        .forEach((pathFragment) => {
          numberOfResponses++;
          if (!checkerFn(removeParenthesis(pathFragment))) {
            numbnerOfFalseResponses++;
            config.report({
              message: `Invalid URL casing`,
              path: path,
              method: "Nil",
            });
          }
        });
    });
  } catch (err) {
    console.log(err);
  }
  const score =
    (Math.max(numberOfResponses - numbnerOfFalseResponses, 0) /
      numberOfResponses) *
    100;
  config.setScore("quality", score);
}
