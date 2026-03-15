#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const repoRoot = path.resolve(__dirname, "..");

const DEFAULT_DIST_DIR = path.join(repoRoot, "dist");
const DEFAULT_OUT_DIR = path.join(DEFAULT_DIST_DIR, "npm");
const ROOT_BIN_SOURCE = path.join(__dirname, "root", "bin", "webtool.js");
const LICENSE_SOURCE = path.join(repoRoot, "LICENSE");
const REPOSITORY_URL = "git+https://github.com/machinae/webtool.git";
const HOMEPAGE_URL = "https://github.com/machinae/webtool";
const BUGS_URL = "https://github.com/machinae/webtool/issues";

const TARGETS = [
  {
    goos: "darwin",
    goarch: "amd64",
    npmOs: "darwin",
    npmCpu: "x64",
    packageNameSuffix: "darwin-x64",
    binaryName: "webtool",
  },
  {
    goos: "darwin",
    goarch: "arm64",
    npmOs: "darwin",
    npmCpu: "arm64",
    packageNameSuffix: "darwin-arm64",
    binaryName: "webtool",
  },
  {
    goos: "linux",
    goarch: "amd64",
    npmOs: "linux",
    npmCpu: "x64",
    packageNameSuffix: "linux-x64",
    binaryName: "webtool",
  },
  {
    goos: "linux",
    goarch: "arm64",
    npmOs: "linux",
    npmCpu: "arm64",
    packageNameSuffix: "linux-arm64",
    binaryName: "webtool",
  },
  {
    goos: "windows",
    goarch: "amd64",
    npmOs: "win32",
    npmCpu: "x64",
    packageNameSuffix: "win32-x64",
    binaryName: "webtool.exe",
  },
  {
    goos: "windows",
    goarch: "arm64",
    npmOs: "win32",
    npmCpu: "arm64",
    packageNameSuffix: "win32-arm64",
    binaryName: "webtool.exe",
  },
];

function parseArgs(argv) {
  const args = {};
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (!arg.startsWith("--")) {
      continue;
    }
    const key = arg.slice(2);
    const value = argv[i + 1];
    if (!value || value.startsWith("--")) {
      args[key] = true;
      continue;
    }
    args[key] = value;
    i += 1;
  }
  return args;
}

function normalizeVersion(rawVersion) {
  if (!rawVersion || typeof rawVersion !== "string") {
    throw new Error("Missing version. Pass --version or provide dist/metadata.json.");
  }
  return rawVersion.startsWith("v") ? rawVersion.slice(1) : rawVersion;
}

async function readJson(filePath) {
  const contents = await fs.readFile(filePath, "utf8");
  return JSON.parse(contents);
}

async function ensureExecutable(filePath) {
  await fs.chmod(filePath, 0o755);
}

async function copyFile(sourcePath, destinationPath) {
  await fs.mkdir(path.dirname(destinationPath), { recursive: true });
  await fs.copyFile(sourcePath, destinationPath);
}

function getBinaryArtifact(artifacts, target) {
  return artifacts.find(
    (artifact) =>
      artifact.type === "Binary" &&
      artifact.extra?.ID === "webtool" &&
      artifact.goos === target.goos &&
      artifact.goarch === target.goarch,
  );
}

function buildCommonPackageJsonFields() {
  return {
    license: "Apache-2.0",
    repository: {
      type: "git",
      url: REPOSITORY_URL,
    },
    homepage: HOMEPAGE_URL,
    bugs: {
      url: BUGS_URL,
    },
    keywords: ["browser", "automation", "chrome", "cli", "cdp", "agent"],
    engines: {
      node: ">=16",
    },
  };
}

function buildRootPackageJson({ packageName, version, optionalDependencies }) {
  return {
    name: packageName,
    version,
    description: "A CLI for your browser.",
    ...buildCommonPackageJsonFields(),
    bin: {
      [packageName]: "bin/webtool.js",
    },
    files: ["bin"],
    optionalDependencies,
  };
}

function buildPlatformPackageJson({ packageName, version, target }) {
  return {
    name: packageName,
    version,
    description: `Platform package for webtool (${target.packageNameSuffix}).`,
    ...buildCommonPackageJsonFields(),
    os: [target.npmOs],
    cpu: [target.npmCpu],
    files: ["vendor"],
  };
}

function buildRootReadme({ packageName }) {
  return `# ${packageName}

Install with:

\`npm install -g ${packageName}\`
`;
}

function buildPlatformReadme({ rootPackageName, platformPackageName }) {
  return `# ${platformPackageName}

This is an internal platform package for ${rootPackageName}. Install the root package instead:

\`npm install -g ${rootPackageName}\`
`;
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  const distDir = path.resolve(args.dist ?? DEFAULT_DIST_DIR);
  const outDir = path.resolve(args.out ?? DEFAULT_OUT_DIR);
  const packageName = args["package-name"] ?? "webtool";

  const metadata = await readJson(path.join(distDir, "metadata.json"));
  const artifacts = await readJson(path.join(distDir, "artifacts.json"));
  const version = normalizeVersion(args.version ?? metadata.version);

  await fs.rm(outDir, { recursive: true, force: true });
  await fs.mkdir(outDir, { recursive: true });

  const optionalDependencies = {};
  const platformPackages = [];

  for (const target of TARGETS) {
    const artifact = getBinaryArtifact(artifacts, target);
    if (!artifact) {
      throw new Error(
        `Missing GoReleaser binary for ${target.goos}/${target.goarch}. Run GoReleaser first.`,
      );
    }

    const platformPackageName = `${packageName}-${target.packageNameSuffix}`;
    const packageDir = path.join(outDir, target.packageNameSuffix);
    const packageJsonPath = path.join(packageDir, "package.json");
    const readmePath = path.join(packageDir, "README.md");
    const licensePath = path.join(packageDir, "LICENSE");
    const binaryPath = path.join(packageDir, "vendor", target.binaryName);

    await copyFile(path.resolve(repoRoot, artifact.path), binaryPath);
    await copyFile(LICENSE_SOURCE, licensePath);
    if (target.npmOs !== "win32") {
      await ensureExecutable(binaryPath);
    }

    const packageJson = buildPlatformPackageJson({
      packageName: platformPackageName,
      version,
      target,
    });
    await fs.writeFile(packageJsonPath, `${JSON.stringify(packageJson, null, 2)}\n`);
    await fs.writeFile(
      readmePath,
      buildPlatformReadme({ rootPackageName: packageName, platformPackageName }),
    );

    optionalDependencies[platformPackageName] = version;
    platformPackages.push({
      name: platformPackageName,
      version,
      dir: path.relative(repoRoot, packageDir),
      os: target.npmOs,
      cpu: target.npmCpu,
      binary: target.binaryName,
    });
  }

  const rootDir = path.join(outDir, "root");
  const rootBinPath = path.join(rootDir, "bin", "webtool.js");
  const rootPackageJsonPath = path.join(rootDir, "package.json");
  const rootReadmePath = path.join(rootDir, "README.md");
  const rootLicensePath = path.join(rootDir, "LICENSE");

  await copyFile(ROOT_BIN_SOURCE, rootBinPath);
  await copyFile(LICENSE_SOURCE, rootLicensePath);
  await ensureExecutable(rootBinPath);
  await fs.writeFile(
    rootPackageJsonPath,
    `${JSON.stringify(buildRootPackageJson({ packageName, version, optionalDependencies }), null, 2)}\n`,
  );
  await fs.writeFile(rootReadmePath, buildRootReadme({ packageName }));

  const manifest = {
    packageName,
    version,
    root: {
      name: packageName,
      version,
      dir: path.relative(repoRoot, rootDir),
    },
    platforms: platformPackages,
    publishOrder: [
      ...platformPackages.map((pkg) => ({
        dir: pkg.dir,
        name: pkg.name,
        version: pkg.version,
      })),
      {
        dir: path.relative(repoRoot, rootDir),
        name: packageName,
        version,
      },
    ],
  };

  await fs.writeFile(path.join(outDir, "manifest.json"), `${JSON.stringify(manifest, null, 2)}\n`);

  console.log(`Generated npm packages for ${packageName}@${version} in ${outDir}`);
}

main().catch((error) => {
  console.error(error instanceof Error ? error.message : error);
  process.exit(1);
});
