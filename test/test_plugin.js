import cmd from "apic/exec";

export default function (cfg, opts) {
  const output = cmd("date");
  console.log("This is executed from test_plugin command", output?.data);
  cfg.setScore("performance", 100);
}
