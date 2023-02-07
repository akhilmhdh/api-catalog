// Kudos: https://github.com/aceakash/string-similarity
function compareTwoStrings(first, second) {
  first = first.replace(/\s+/g, "");
  second = second.replace(/\s+/g, "");

  if (first === second) return 1; // identical or empty
  if (first.length < 2 || second.length < 2) return 0; // if either is a 0-letter or 1-letter string

  let firstBigrams = new Map();
  for (let i = 0; i < first.length - 1; i++) {
    const bigram = first.substring(i, i + 2);
    const count = firstBigrams.has(bigram) ? firstBigrams.get(bigram) + 1 : 1;

    firstBigrams.set(bigram, count);
  }

  let intersectionSize = 0;
  for (let i = 0; i < second.length - 1; i++) {
    const bigram = second.substring(i, i + 2);
    const count = firstBigrams.has(bigram) ? firstBigrams.get(bigram) : 0;

    if (count > 0) {
      firstBigrams.set(bigram, count - 1);
      intersectionSize++;
    }
  }

  return (2.0 * intersectionSize) / (first.length + second.length - 2);
}

function stripOfBaseURL(path, baseURLs) {
  for (let i = 0; i < baseURLs.length; i++) {
    if (path.startsWith(baseURLs[i])) {
      return path.slice(baseURLs[i].length);
    }
  }
  return path;
}
export default function (config, options) {
  let numberOfResponses = 0;
  let numbnerOfFalseResponses = 0;

  const weight = options?.weight || 0.8;
  const blackListPaths = options?.blacklist_paths || [];
  const baseURLs = options?.base_urls || [];

  const paths = Object.keys(config.schema.paths);
  for (let i = 0; i < paths.length; i++) {
    // skip blacklisted paths
    const pathA = stripOfBaseURL(paths[i], baseURLs);

    if (blackListPaths.includes(paths[i])) continue;
    for (let j = paths.length - 1; j >= i; j--) {
      numberOfResponses++;
      if (j !== i) {
        const pathB = stripOfBaseURL(paths[j], baseURLs);

        const similiarity = compareTwoStrings(pathA, pathB);
        if (similiarity > weight) {
          numbnerOfFalseResponses++;

          // get all methods
          const methods = Object.keys(config.schema.paths[paths[i]])
            .join(", ")
            .toUpperCase();

          config.report({
            message: `URL ${paths[i]} similiar to ${paths[j]}, similiarity: ${similiarity}`,
            path: paths[i],
            method: methods,
          });
        }
      }
    }
  }

  const score =
    (Math.max(numberOfResponses - numbnerOfFalseResponses, 0) /
      numberOfResponses) *
    100;

  config.setScore("quality", score);
}
