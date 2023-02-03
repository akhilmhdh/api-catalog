// By this spec: https://perishablepress.com/stop-using-unsafe-characters-in-urls/
const unsafeURLRegex = /^[a-zA-Z0-9{}\/~_-]*$/;

export default function (config) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;

  Object.keys(config.schema.paths).forEach((path) => {
    numberOfResponses++;
    if (!unsafeURLRegex.test(path)) {
      numbnerOfFalseResponses++;

      // get all methods
      const methods = Object.keys(config.schema.paths[path])
        .join(", ")
        .toUpperCase();
      config.report({
        message: `URL contains unsafe character`,
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
