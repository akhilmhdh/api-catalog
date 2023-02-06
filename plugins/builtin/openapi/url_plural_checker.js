import { isPlural, isSingular } from "apic/strings";

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

  const type = options?.type || "singular";
  const blackListPaths = options?.blacklist_paths || [];

  const checkerFn = type === "singular" ? isSingular : isPlural;

  Object.keys(config.schema.paths).forEach((path) => {
    // next iteration
    if (blackListPaths.includes(path)) return;

    const pathFragment = path.split("/").filter(Boolean);

    for (let i = 0; i < pathFragment.length; i++) {
      // dont need to check dynamic params like /pets/{petID} -> petID is just a variable
      if (isDynamicParams(pathFragment[i])) continue;
      // then if its last pathFragment and is having an extension like .json  those are files
      if (i === pathFragment.length - 1 && pathFragment[i].includes("."))
        continue;

      numberOfResponses++;
      // check casing
      if (!checkerFn(pathFragment[i])) {
        numbnerOfFalseResponses++;
        // get all methods
        const methods = Object.keys(config.schema.paths[path])
          .join(", ")
          .toUpperCase();

        config.report({
          message: `URL is is not ${type}`,
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
