// Collects the license of each npm package included in the bundle.
// webpack.config.js writes them at the top of the bundle after minification,
// because the minifier drops comments inside some packages (e.g. @xterm/xterm).
// Based on license-loader 0.5.0 by Cole Chamberlain (MIT), with support for
// scoped packages such as @xterm/xterm.

const path = require('path');
const fs = require('fs');

const licenseFileNames = ['LICENSE', 'LICENCE', 'LICENSE.md', 'LICENCE.md', 'LICENSE.txt', 'LICENCE.txt',
    'license', 'licence', 'license.md', 'licence.md', 'license.txt', 'licence.txt'];

const seen = new Set();
const licenses = [];

module.exports = function (source) {
    this.cacheable && this.cacheable();
    const parts = this.resourcePath.split(path.sep);
    const idx = parts.lastIndexOf('node_modules');
    if (idx === -1) {
        return source;
    }
    const nameLength = parts[idx + 1].startsWith('@') ? 2 : 1;
    const packageDir = parts.slice(0, idx + 1 + nameLength).join(path.sep);
    const packageJSON = JSON.parse(fs.readFileSync(path.join(packageDir, 'package.json'), 'utf8'));

    const fingerprint = `${packageJSON.name}:${packageJSON.version}`;
    if (seen.has(fingerprint)) {
        return source;
    }
    seen.add(fingerprint);

    const licenseFile = licenseFileNames.find((f) => fs.existsSync(path.join(packageDir, f)));
    const licenseText = licenseFile ? fs.readFileSync(path.join(packageDir, licenseFile), 'utf8').trimEnd() : '';
    const header = [
        `${packageJSON.name} (https://npmjs.com/package/${packageJSON.name})`,
        `@license ${packageJSON.license || ''}`.trimEnd(),
        `@version ${packageJSON.version}`,
    ];
    if (licenseText) {
        header.push(`==${packageJSON.name}/${licenseFile}==`, ...licenseText.replace(/\*\//g, '').split('\n'));
    }
    licenses.push(`/*!\n${header.map((l) => ` * ${l}`.trimEnd()).join('\n')}\n */`);
    return source;
};

module.exports.licenses = licenses;
