const { createHash } = require('crypto');
const { execSync } = require('child_process');
const fs = require('fs');
const https = require('https');
const http = require('http');
const path = require('path');
const os = require('os');

const REPO = 'gogodjzhu/word-flow';
const PLATFORM_MAP = {
  darwin: { x64: 'darwin-amd64', arm64: 'darwin-arm64' },
  linux: { x64: 'linux-amd64', arm64: 'linux-arm64' },
  win32: { x64: 'windows-amd64', arm64: 'windows-arm64' },
};

function getPackageVersion() {
  const pkg = JSON.parse(fs.readFileSync(path.join(__dirname, 'package.json'), 'utf8'));
  return pkg.version;
}

function getTarget() {
  const platform = process.platform;
  const arch = process.arch;
  if (!PLATFORM_MAP[platform] || !PLATFORM_MAP[platform][arch]) {
    console.error(`Unsupported platform/arch: ${platform}/${arch}`);
    console.error('Supported combinations: darwin/x64, darwin/arm64, linux/x64, linux/arm64, win32/x64, win32/arm64');
    process.exit(1);
  }
  return PLATFORM_MAP[platform][arch];
}

function getBinaryName(target) {
  if (target.startsWith('windows')) {
    return 'wordflow.exe';
  }
  return 'wordflow';
}

function getArchiveName(target) {
  const ext = target.startsWith('windows') ? 'zip' : 'tar.gz';
  return `wordflow-${target}.${ext}`;
}

function download(url, expectContentType) {
  return new Promise((resolve, reject) => {
    const client = url.startsWith('https') ? https : http;
    const follow = (u, redirects = 0) => {
      if (redirects > 10) {
        return reject(new Error('Too many redirects'));
      }
      client.get(u, { headers: { 'User-Agent': 'wordflow-installer' } }, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return follow(res.headers.location, redirects + 1);
        }
        if (res.statusCode !== 200) {
          res.resume();
          return reject(new Error(`HTTP ${res.statusCode} for ${u}`));
        }
        const contentType = (res.headers['content-type'] || '').split(';')[0].trim();
        if (expectContentType && contentType && !contentType.startsWith(expectContentType)) {
          res.resume();
          if (contentType.startsWith('text/html')) {
            return reject(new Error(`Release not found at ${u}. Make sure the GitHub release v<version> exists.`));
          }
          return reject(new Error(`Unexpected content type: ${contentType} (expected ${expectContentType})`));
        }
        const chunks = [];
        res.on('data', (chunk) => chunks.push(chunk));
        res.on('end', () => resolve(Buffer.concat(chunks)));
        res.on('error', reject);
      }).on('error', reject);
    };
    follow(url);
  });
}

function sha256(buffer) {
  return createHash('sha256').update(buffer).digest('hex');
}

function extractArchive(archiveBuffer, target, binaryName) {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'wordflow-'));

  if (target.startsWith('windows')) {
    const admName = 'adm-zip';
    try {
      if (!fs.existsSync(path.join(__dirname, 'node_modules', admName))) {
        execSync(`npm install --no-save ${admName}`, { cwd: __dirname, stdio: 'pipe' });
      }
      const AdmZip = require(admName);
      const zip = new AdmZip(archiveBuffer);
      zip.extractAllTo(tmpDir, true);
    } finally {
      const admPath = path.join(__dirname, 'node_modules', admName);
      if (fs.existsSync(admPath)) {
        fs.rmSync(admPath, { recursive: true, force: true });
      }
    }
  } else {
    const archivePath = path.join(tmpDir, 'archive.tar.gz');
    fs.writeFileSync(archivePath, archiveBuffer);
    execSync(`tar -xzf "${archivePath}" -C "${tmpDir}"`, { stdio: 'pipe' });
    fs.unlinkSync(archivePath);
  }

  const findBinary = (dir) => {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);
      if (entry.isFile() && entry.name === binaryName) {
        return fullPath;
      }
      if (entry.isDirectory()) {
        const result = findBinary(fullPath);
        if (result) return result;
      }
    }
    return null;
  };

  const binaryPath = findBinary(tmpDir);
  if (!binaryPath) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
    throw new Error(`Binary '${binaryName}' not found in archive`);
  }

  const destPath = path.join(__dirname, binaryName);
  fs.copyFileSync(binaryPath, destPath);
  fs.rmSync(tmpDir, { recursive: true, force: true });
  return destPath;
}

async function main() {
  if (process.env.WORDFLOW_SKIP_INSTALL) {
    console.log('Skipping wordflow binary download (WORDFLOW_SKIP_INSTALL is set)');
    return;
  }

  const binaryOverride = process.env.WORDFLOW_BINARY_PATH;
  if (binaryOverride) {
    console.log(`Using wordflow binary from WORDFLOW_BINARY_PATH: ${binaryOverride}`);
    return;
  }

  const version = getPackageVersion();
  const target = getTarget();
  const binaryName = getBinaryName(target);
  const archiveName = getArchiveName(target);
  const baseUrl = `https://github.com/${REPO}/releases/download/v${version}`;
  const archiveUrl = `${baseUrl}/${archiveName}`;
  const checksumsUrl = `${baseUrl}/SHA256SUMS`;

  const destBinary = path.join(__dirname, binaryName);

  if (fs.existsSync(destBinary)) {
    console.log(`wordflow binary already exists at ${destBinary}, skipping download`);
    return;
  }

  console.log(`Installing wordflow v${version} for ${target}...`);
  console.log(`Downloading ${archiveUrl}...`);

  let archiveBuffer;
  try {
    archiveBuffer = await download(archiveUrl, 'application/');
  } catch (err) {
    console.error(`Failed to download wordflow binary: ${err.message}`);
    console.error('');
    console.error('Manual installation instructions:');
    console.error(`  1. Download ${archiveUrl}`);
    console.error(`  2. Extract '${binaryName}' and place it at: ${destBinary}`);
    console.error('');
    console.error('Or build from source:');
    console.error('  git clone https://github.com/gogodjzhu/word-flow.git');
    console.error('  cd word-flow && go build -o wordflow cmd/wordflow/main.go');
    console.error(`  cp wordflow ${destBinary}`);
    process.exit(1);
  }

  console.log('Verifying checksum...');
  let checksums;
  try {
    const checksumsBuffer = await download(checksumsUrl, 'text/');
    checksums = checksumsBuffer.toString('utf8');
  } catch (err) {
    console.error(`Warning: Could not download checksums file: ${err.message}`);
    console.error('Skipping checksum verification (binary integrity cannot be verified)');
  }

  const archiveHash = sha256(archiveBuffer);
  if (checksums) {
    const expectedLine = `${archiveHash}  ${archiveName}`;
    if (!checksums.includes(expectedLine) && !checksums.includes(`${archiveHash} *${archiveName}`)) {
      const lines = checksums.trim().split('\n');
      const matchedLine = lines.find((l) => l.endsWith(archiveName) || l.endsWith(`*/${archiveName}`));
      if (!matchedLine || !matchedLine.startsWith(archiveHash)) {
        console.error(`Checksum verification failed!`);
        console.error(`Expected: ${matchedLine}`);
        console.error(`Actual:   ${archiveHash}  ${archiveName}`);
        process.exit(1);
      }
    }
    console.log('Checksum verified.');
  }

  console.log('Extracting binary...');
  try {
    const extractedPath = extractArchive(archiveBuffer, target, binaryName);
    if (process.platform !== 'win32') {
      fs.chmodSync(extractedPath, 0o755);
    }
    console.log(`wordflow v${version} installed successfully at ${extractedPath}`);
  } catch (err) {
    console.error(`Failed to extract binary: ${err.message}`);
    process.exit(1);
  }
}

main().catch((err) => {
  console.error(`Installation failed: ${err.message}`);
  process.exit(1);
});