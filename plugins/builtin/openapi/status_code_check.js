export default function (config) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;

  Object.keys(config.schema.paths).forEach((path) => {
    Object.keys(config.schema.paths[path]).forEach((method) => {
      Object.keys(config.schema.paths[path][method].responses).forEach(
        (responseStatusCode) => {
          numberOfResponses++;
          // convert string to number for statuscode
          const code = parseInt(responseStatusCode, 10);
          if (responseStatusCode !== "default" && Number.isNaN(code)) {
            numbnerOfFalseResponses++;
            config.report({
              message: `Invalid status code - ${responseStatusCode}`,
              path: path,
              method: method,
            });
          } else if (
            responseStatusCode !== "default" &&
            (code < 100 || code > 599)
          ) {
            numbnerOfFalseResponses++;
            config.report({
              message: `Invalid status code - ${responseStatusCode}`,
              path: path,
              method: method,
            });
          }
        }
      );
    });
  });
  const score =
    ((numberOfResponses - numbnerOfFalseResponses) / numberOfResponses) * 100;
  config.setScore("quality", score);
}
