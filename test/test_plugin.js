import cmd from "apic/exec";

export default function (cfg, opts) {
  try {
    const output = cmd("date");
    console.log("hello world", opts?.test_data);
  } catch (err) {
    console.log(err);
  }
}
