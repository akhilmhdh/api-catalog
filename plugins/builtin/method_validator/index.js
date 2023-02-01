export default function (config) {
  console.log("header method validator");
  config.setScore("security", 92.8);
  config.report({
    message: "hello world",
    path: "/pets",
    method: "POST",
    headers: [{ key: "x-http", value: "something" }],
    metadata: {
      ref: "something random shit",
    },
  });
}
