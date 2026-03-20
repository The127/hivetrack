#!/usr/bin/env node
"use strict";

const { spawn } = require("child_process");
const path = require("path");

const ext = process.platform === "win32" ? ".exe" : "";
const binPath = path.join(__dirname, `hivetrack-mcp${ext}`);

const child = spawn(binPath, process.argv.slice(2), {
  stdio: "inherit",
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code);
  }
});
