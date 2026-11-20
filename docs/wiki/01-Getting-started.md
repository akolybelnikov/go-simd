# Getting started

Use the local browser to see the available operations and run a first comparison. For remote sessions or automation, the terminal commands run the same experiments.

## Open the lab

If you downloaded an archive, extract it and open a terminal in that folder.

| Platform | Command |
| --- | --- |
| Windows PowerShell | `.\simdlab.exe ui` |
| Linux | `./simdlab ui` |
| Executable on your PATH | `simdlab ui` |

If you are building from source, follow the [installation guide](https://github.com/akolybelnikov/go-simd/blob/main/docs/cli.md#install-the-cli-from-source).

The browser opens on **Explore**. Choose **Addition**, review the methods and values per input array, and select **Preview comparison**. In **Compare**, choose **Run locally**. **Results** shows the median timing difference and individual repeats.

If the browser does not open, copy the private address printed in the terminal. Leave the terminal running during the session. Ctrl+C stops the lab.

## Continue learning

- [Run and interpret your first comparison](02-First-experiment).
- [Browse the methods and what they measure](32-Operation-measurement-guide).
- [Save and reopen your report](05-Reports-and-evidence).
