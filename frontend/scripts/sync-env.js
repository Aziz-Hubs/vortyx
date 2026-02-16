const fs = require('fs');
const path = require('path');

const rootEnvPath = path.join(__dirname, '..', '..', '.env');
const frontendEnvPath = path.join(__dirname, '..', '.env.local');

const rootEnv = fs.readFileSync(rootEnvPath, 'utf-8');

const lines = rootEnv.split('\n');
const varsToInclude = [
  'ZITADEL_DOMAIN',
  'ZITADEL_ISSUER',
  'ZITADEL_CLIENT_ID',
  'ZITADEL_CLIENT_SECRET',
  'ZITADEL_MANAGEMENT_PAT',
  'ZITADEL_PROJECT_ID',
  'ZITADEL_AUDIENCES',
  'NEXTAUTH_URL',
  'NEXTAUTH_SECRET',
  'AUTH_SECRET',
  'NEXT_PUBLIC_API_URL',
];

const envVars = {};

let output = '# Auto-generated from root .env - do not edit manually\n';

for (const line of lines) {
  const trimmed = line.trim();
  if (trimmed.startsWith('#') || trimmed === '') continue;
  
  const [key, ...valueParts] = trimmed.split('=');
  if (!key) continue;
  
  const value = valueParts.join('=').trim();
  envVars[key] = value;
  
  if (varsToInclude.includes(key)) {
    output += `${key}=${value}\n`;
  }
}

output += '\n# Frontend-only vars (derived from root .env)\n';
output += `NEXT_PUBLIC_ZITADEL_API=${envVars['ZITADEL_ISSUER'] || envVars['ZITADEL_DOMAIN'] || ''}\n`;

fs.writeFileSync(frontendEnvPath, output);

console.log('Synced environment variables from root .env to frontend/.env.local');
