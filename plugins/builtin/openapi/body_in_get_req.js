export default function (config) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;

  Object.keys(config.schema.paths).forEach((path) => {
    Object.keys(config.schema.paths[path]).forEach((method) => {
      if (method === "get") {
        numberOfResponses++;
        if (config.schema.paths[path][method].hasOwnProperty("requestBody")) {
          numbnerOfFalseResponses++;
          config.report({
            message: "Request body in GET request",
            path: path,
            method: method,
          });
        }

        (config.schema.paths[path][method].parameters || []).forEach(
          (param) => {
            if (param.in === "body") {
              numbnerOfFalseResponses++;
              config.report({
                message: "Request body in GET request",
                path: path,
                method: method,
              });
            }
          }
        );
      }
    });
  });
  // if number goes to negative
  const score =
    (Math.max(numberOfResponses - numbnerOfFalseResponses, 0) /
      numberOfResponses) *
    100;
  config.setScore("security", score);
}
