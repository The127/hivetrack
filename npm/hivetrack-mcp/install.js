#!/usr/bin/env node
"use strict";

const { execSync } = require("child_process");
const fs = require("fs");
const https = require("https");
const path = require("path");

const REPO = "The127/hivetrack";
const BINARY = "hivetrack-mcp";

const PLATFORM_MAP = {
  linux: "linux",
  darwin: "darwin",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const os = PLATFORM_MAP[process.platform];
const arch = ARCH_MAP[process.arch];

if (!os || !arch) {
  console.error(
    `Unsupported platform: ${process.platform}/${process.arch}`
  );
  process.exit(1);
}

const ext = os === "windows" ? ".exe" : "";
const assetName = `${BINARY}-${os}-${arch}${ext}`;
const binDir = __dirname;
const binPath = path.join(binDir, `${BINARY}${ext}`);

function fetch(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, { headers: { "User-Agent": "hivetrack-mcp-npm" } }, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return fetch(res.headers.location).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
        }
        const chunks = [];
        res.on("data", (c) => chunks.push(c));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

async function main() {
  const pkgVersion = require("./package.json").version;
  const tag = `v${pkgVersion}`;

  console.log(`Installing ${BINARY} ${tag} (${os}/${arch})...`);

  const baseUrl = `https://github.com/${REPO}/releases/download/${tag}`;

  // Download binary
  const binary = await fetch(`${baseUrl}/${assetName}`);

  // Download checksums and verify
  const checksumsData = await fetch(`${baseUrl}/checksums.txt`);
  const checksums = checksumsData.toString();
  const line = checksums.split("\n").find((l) => l.includes(assetName));
  if (line) {
    const expected = line.split(/\s+/)[0];
    const crypto = require("crypto");
    const actual = crypto.createHash("sha256").update(binary).digest("hex");
    if (expected !== actual) {
      console.error(`Checksum mismatch: expected ${expected}, got ${actual}`);
      process.exit(1);
    }
    console.log("Checksum verified.");
  }

  fs.writeFileSync(binPath, binary);
  fs.chmodSync(binPath, 0o755);
  console.log(`Installed to ${binPath}`);
}

main().catch((err) => {
  console.error(`Failed to install ${BINARY}: ${err.message}`);
  process.exit(1);
});
