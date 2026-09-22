const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

try {
  require.resolve('terser');
} catch (e) {
  console.log('Installing terser...');
  execSync('npm install terser', { stdio: 'inherit', cwd: __dirname });
}

const terser = require('terser');

const publicDir = path.join(__dirname, 'public');
const distDir = path.join(__dirname, 'public', 'dist');

if (!fs.existsSync(distDir)) {
  fs.mkdirSync(distDir);
}

const filesToMinify = [
  'app-core.js',
  'app-extra.js',
  'app-penilaian.js',
  'app-sangu.js',
  'app-tahfidz.js',
  'app-views.js',
  'app-absensi.js',
  'app-nilai-kegiatan.js'
];

async function minifyFiles() {
  for (const file of filesToMinify) {
    const filePath = path.join(publicDir, file);
    if (!fs.existsSync(filePath)) continue;
    
    console.log(`Minifying ${file}...`);
    const code = fs.readFileSync(filePath, 'utf8');
    
    const minified = await terser.minify(code, {
      mangle: { toplevel: false },
      compress: {
        drop_console: true,
        passes: 2
      }
    });
    
    if (minified.error) {
      console.error(`Error minifying ${file}:`, minified.error);
      continue;
    }
    
    const distPath = path.join(distDir, file.replace('.js', '-v4.min.js'));
    fs.writeFileSync(distPath, minified.code);
    console.log(`✅ Saved ${distPath}`);
  }
}

minifyFiles().catch(console.error);
