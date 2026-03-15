#!/usr/bin/env node

const { spawn } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const ROOT_PACKAGE_NAME = "webtool";
const TARGETS = {
  darwin: {
    arm64: { packageName: "webtool-darwin-arm64", binary: "webtool" },
    x64: { packageName: "webtool-darwin-x64", binary: "webtool" },
  },
  linux: {
    arm64: { packageName: "webtool-linux-arm64", binary: "webtool" },
    x64: { packageName: "webtool-linux-x64", binary: "webtool" },
  },
  win32: {
    arm64: { packageName: "webtool-win32-arm64", binary: "webtool.exe" },
    x64: { packageName: "webtool-win32-x64", binary: "webtool.exe" },
  },
};

function installHint() {
  return `npm install -g ${ROOT_PACKAGE_NAME}@latest`;
}

function resolveBinary() {
  const platformTargets = TARGETS[process.platform];
  if (!platformTargets) {
    throw new Error(`Unsupported platform: ${process.platform}`);
  }

  const target = platformTargets[process.arch];
  if (!target) {
    throw new Error(`Unsupported architecture: ${process.platform}/${process.arch}`);
  }

  let packageJsonPath;
  try {
    packageJsonPath = require.resolve(`${target.packageName}/package.json`);
  } catch {
    throw new Error(
      `Missing optional dependency ${target.packageName}. Reinstall webtool: ${installHint()}`,
    );
  }

  const packageRoot = path.dirname(packageJsonPath);
  const binaryPath = path.join(packageRoot, "vendor", target.binary);
  if (!fs.existsSync(binaryPath)) {
    throw new Error(
      `Missing binary for ${target.packageName} at ${binaryPath}. Reinstall webtool: ${installHint()}`,
    );
  }

  return binaryPath;
}

const binaryPath = resolveBinary();
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  env: process.env,
});

child.on("error", (err) => {
  console.error(err);
  process.exit(1);
});

const forwardSignal = (signal) => {
  if (child.killed) {
    return;
  }
  try {
    child.kill(signal);
  } catch {
    // Ignore errors if the child has already exited.
  }
};

["SIGINT", "SIGTERM", "SIGHUP"].forEach((signal) => {
  process.on(signal, () => forwardSignal(signal));
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 1);
});
