export default function (config, options) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;

  const allowedStatusCodes = options?.allowed_status_codes;
  Object.keys(config.schema.paths).forEach((path) => {
    Object.keys(config.schema.paths[path]).forEach((method) => {
      Object.keys(config.schema.paths[path][method].responses).forEach(
        (responseStatusCode) => {
          numberOfResponses++;
          // convert string to number for statuscode
          const code = parseInt(responseStatusCode, 10);
          if (responseStatusCode !== "default") {
            if (Number.isNaN(code)) {
              numbnerOfFalseResponses++;
              config.report({
                message: `Invalid status code - ${responseStatusCode}`,
                path: path,
                method: method,
              });
            } else if (code < 100 || code > 599) {
              numbnerOfFalseResponses++;
              config.report({
                message: `Invalid status code - ${responseStatusCode}`,
                path: path,
                method: method,
              });
            } else if (
              Boolean(allowedStatusCodes) &&
              !allowedStatusCodes.includes(responseStatusCode)
            ) {
              numbnerOfFalseResponses++;
              config.report({
                message: `Statuscode is not allowed - ${responseStatusCode}`,
                path: path,
                method: method,
              });
            }
          }
        }
      );
    });
  });
  const score =
    ((numberOfResponses - numbnerOfFalseResponses) / numberOfResponses) * 100;
  config.setScore("quality", score);
}
