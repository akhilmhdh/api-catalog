import { isCasing } from "apic/strings";

// for dynamic parameters like /pets/{something}
function isDynamicParams(path) {
  if (path[0] === "{" && path[path.length - 1] === "}") {
    return true;
  }

  return false;
}

export default function (config, options) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;

  const casing = options?.casing || "camelcase";

  Object.keys(config.schema.paths).forEach((path) => {
    const pathFragment = path.split("/").filter(Boolean);

    for (let i = 0; i < pathFragment.length; i++) {
      // dont need to check dynamic params like /pets/{petID} -> petID is just a variable
      if (isDynamicParams(pathFragment[i])) continue;
      // then if its last pathFragment and is having an extension like .json  those are files
      if (i === pathFragment.length - 1 && pathFragment[i].includes("."))
        continue;

      numberOfResponses++;
      // check casing
      if (!isCasing(casing, pathFragment[i])) {
        numbnerOfFalseResponses++;
        // get all methods
        const methods = Object.keys(config.schema.paths[path])
          .join(", ")
          .toUpperCase();

        config.report({
          message: `Invalid URL casing`,
          path: path,
          method: methods,
        });
      }
    }
  });
  console.log(numberOfResponses, numbnerOfFalseResponses);

  const score =
    (Math.max(numberOfResponses - numbnerOfFalseResponses, 0) /
      numberOfResponses) *
    100;
  config.setScore("quality", score);
}
