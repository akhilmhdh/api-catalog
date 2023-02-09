/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    "introduction",
    "architecture",
    {
      type: "category",
      label: "CLI",
      items: [
        "cli/overview",
        "cli/installation",
        {
          type: "category",
          label: "Rules",
          items: ["cli/rules/what-are-rules", "cli/rules/customizing"],
        },
        {
          type: "category",
          label: "Builtin Rules",
          items: [
            "cli/rules/builtin/overview",
            {
              type: "category",
              label: "OpenAPI",
              items: ["cli/rules/builtin/openapi/url-length-check"],
            },
          ],
        },
        {
          type: "category",
          label: "User Defined Rules",
          items: [
            "cli/rules/user-defined/overview",
            "cli/rules/user-defined/our-first-rule",
          ],
        },
        {
          type: "category",
          label: "Modules",
          items: [
            "cli/modules/overview",
            "cli/modules/exec",
            "cli/modules/env",
          ],
        },
        {
          type: "category",
          label: "CLI Commands",
          items: ["cli/cli-commands/apic-run", "cli/cli-commands/apic-help"],
        },
      ],
    },
    "kudos",
  ],
};

module.exports = sidebars;
