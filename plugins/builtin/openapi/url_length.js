// for giving a score to {something} kinda paths in openapi
function isDynamicPathFragment(path) {
  return path[0] === "{" && path[path.length - 1] === "}";
}

export default function (config, options) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;

  const dynamicPathWeight = options?.weight || 5;
  const maxURLAllowedLength = options?.max_url_length || 75;

  Object.keys(config.schema.paths).forEach((path) => {
    numberOfResponses++;
    const resources = path.split("/").filter(Boolean);
    // the idea is for a given path /pets/{petid}
    // we will ask user a weight for dynamic params then length of path fragment plus the dynamic x weigth
    // gives total length
    const resourceLength = resources.reduce(
      (prev, curr) =>
        isDynamicPathFragment(curr)
          ? prev + dynamicPathWeight
          : prev + curr.length,
      0
    );

    if (resourceLength > maxURLAllowedLength || resources > 10) {
      numbnerOfFalseResponses++;

      // get all methods
      const methods = Object.keys(config.schema.paths[path])
        .join(", ")
        .toUpperCase();
      config.report({
        message: `URL is too big, Resources: ${resources} Length: ${resourceLength} Weight: ${dynamicPathWeight}`,
        path: path,
        method: methods,
      });
    }
  });

  const score =
    (Math.max(numberOfResponses - numbnerOfFalseResponses, 0) /
      numberOfResponses) *
    100;

  config.setScore("quality", score);
}
