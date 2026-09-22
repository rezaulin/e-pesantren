const fs = require('fs');
const p = require('path');
const d = 'd:/abi/backup pesantren/pesantren v20/public';

fs.readdirSync(d).filter(f => f.endsWith('.js')).forEach(f => {
  let c = fs.readFileSync(p.join(d, f), 'utf8');
  let regex = /<button class=\"back-btn\"[^>]*>.*?<\/button>\s*(?:<div class=\"page-header[^>]*>|<div class=\"card au\">)/g;
  let matches = c.match(regex);
  if (matches) {
    console.log(f, matches.length);
  }
});
