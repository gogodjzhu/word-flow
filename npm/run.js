#!/usr/bin/env node

const { execFileSync } = require('child_process');
const path = require('path');

const binaryName = process.platform === 'win32' ? 'wordflow.exe' : 'wordflow';

const binaryPath = process.env.WORDFLOW_BINARY_PATH
  || path.join(__dirname, binaryName);

const args = process.argv.slice(2);

try {
  execFileSync(binaryPath, args, {
    stdio: 'inherit',
    env: { ...process.env },
  });
  process.exit(0);
} catch (err) {
  if (err.status != null) {
    process.exit(err.status);
  }
  console.error(`wordflow: failed to execute binary at ${binaryPath}`);
  console.error(err.message);
  process.exit(1);
}