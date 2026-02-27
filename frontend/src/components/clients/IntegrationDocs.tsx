import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { Client } from '@/types/client';

interface IntegrationDocsProps {
  client: Client;
}

export function IntegrationDocs({ client }: IntegrationDocsProps) {
  const isConfidential = client.client_type === 'confidential';
  const redirectUri = client.redirect_uris[0] || 'https://your-app.com/callback';

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          Integration Guide
          <Badge variant="outline">{client.client_type}</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue={isConfidential ? 'auth-code' : 'pkce'} className="w-full">
          <TabsList className="mb-4 w-full justify-start">
            {isConfidential ? (
              <TabsTrigger value="auth-code">Authorization Code</TabsTrigger>
            ) : (
              <TabsTrigger value="pkce">PKCE Flow</TabsTrigger>
            )}
            <TabsTrigger value="token">Token Exchange</TabsTrigger>
            <TabsTrigger value="refresh">Token Refresh</TabsTrigger>
          </TabsList>

          {/* Authorization Code Flow (confidential) */}
          {isConfidential && (
            <TabsContent value="auth-code" className="space-y-4">
              <p className="text-sm text-muted-foreground">
                Standard Authorization Code flow for server-side applications.
              </p>

              <div>
                <p className="text-sm font-medium mb-2">1. Redirect user to authorize</p>
                <CodeBlock language="cURL">
{`GET /oauth2/auth?\\
  response_type=code&\\
  client_id=${client.client_id}&\\
  redirect_uri=${encodeURIComponent(redirectUri)}&\\
  scope=openid profile email&\\
  state=random_state_value`}
                </CodeBlock>
              </div>

              <div>
                <p className="text-sm font-medium mb-2">2. Exchange code for tokens</p>
                <Tabs defaultValue="curl" className="w-full">
                  <TabsList className="h-8">
                    <TabsTrigger value="curl" className="text-xs">cURL</TabsTrigger>
                    <TabsTrigger value="js" className="text-xs">JavaScript</TabsTrigger>
                    <TabsTrigger value="go" className="text-xs">Go</TabsTrigger>
                  </TabsList>
                  <TabsContent value="curl">
                    <CodeBlock language="bash">
{`curl -X POST /oauth2/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "redirect_uri=${redirectUri}" \\
  -d "client_id=${client.client_id}" \\
  -d "client_secret=YOUR_CLIENT_SECRET"`}
                    </CodeBlock>
                  </TabsContent>
                  <TabsContent value="js">
                    <CodeBlock language="javascript">
{`const response = await fetch('/oauth2/token', {
  method: 'POST',
  headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  body: new URLSearchParams({
    grant_type: 'authorization_code',
    code: authorizationCode,
    redirect_uri: '${redirectUri}',
    client_id: '${client.client_id}',
    client_secret: 'YOUR_CLIENT_SECRET',
  }),
});
const tokens = await response.json();`}
                    </CodeBlock>
                  </TabsContent>
                  <TabsContent value="go">
                    <CodeBlock language="go">
{`data := url.Values{
    "grant_type":    {"authorization_code"},
    "code":          {authorizationCode},
    "redirect_uri":  {"${redirectUri}"},
    "client_id":     {"${client.client_id}"},
    "client_secret": {"YOUR_CLIENT_SECRET"},
}
resp, err := http.PostForm("/oauth2/token", data)`}
                    </CodeBlock>
                  </TabsContent>
                </Tabs>
              </div>
            </TabsContent>
          )}

          {/* PKCE Flow (public) */}
          {!isConfidential && (
            <TabsContent value="pkce" className="space-y-4">
              <p className="text-sm text-muted-foreground">
                PKCE (Proof Key for Code Exchange) flow for public clients (SPAs, mobile apps).
              </p>

              <div>
                <p className="text-sm font-medium mb-2">1. Generate PKCE parameters</p>
                <CodeBlock language="javascript">
{`// Generate code verifier (43-128 chars, URL-safe)
const codeVerifier = generateRandomString(64);
// Create code challenge (SHA256 + base64url)
const codeChallenge = base64url(sha256(codeVerifier));`}
                </CodeBlock>
              </div>

              <div>
                <p className="text-sm font-medium mb-2">2. Redirect user to authorize with PKCE</p>
                <CodeBlock language="cURL">
{`GET /oauth2/auth?\\
  response_type=code&\\
  client_id=${client.client_id}&\\
  redirect_uri=${encodeURIComponent(redirectUri)}&\\
  scope=openid profile email&\\
  state=random_state_value&\\
  code_challenge=CODE_CHALLENGE&\\
  code_challenge_method=S256`}
                </CodeBlock>
              </div>

              <div>
                <p className="text-sm font-medium mb-2">3. Exchange code with code_verifier</p>
                <Tabs defaultValue="curl" className="w-full">
                  <TabsList className="h-8">
                    <TabsTrigger value="curl" className="text-xs">cURL</TabsTrigger>
                    <TabsTrigger value="js" className="text-xs">JavaScript</TabsTrigger>
                    <TabsTrigger value="go" className="text-xs">Go</TabsTrigger>
                  </TabsList>
                  <TabsContent value="curl">
                    <CodeBlock language="bash">
{`curl -X POST /oauth2/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "redirect_uri=${redirectUri}" \\
  -d "client_id=${client.client_id}" \\
  -d "code_verifier=YOUR_CODE_VERIFIER"`}
                    </CodeBlock>
                  </TabsContent>
                  <TabsContent value="js">
                    <CodeBlock language="javascript">
{`const response = await fetch('/oauth2/token', {
  method: 'POST',
  headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  body: new URLSearchParams({
    grant_type: 'authorization_code',
    code: authorizationCode,
    redirect_uri: '${redirectUri}',
    client_id: '${client.client_id}',
    code_verifier: codeVerifier,
  }),
});
const tokens = await response.json();`}
                    </CodeBlock>
                  </TabsContent>
                  <TabsContent value="go">
                    <CodeBlock language="go">
{`data := url.Values{
    "grant_type":     {"authorization_code"},
    "code":           {authorizationCode},
    "redirect_uri":   {"${redirectUri}"},
    "client_id":      {"${client.client_id}"},
    "code_verifier":  {codeVerifier},
}
resp, err := http.PostForm("/oauth2/token", data)`}
                    </CodeBlock>
                  </TabsContent>
                </Tabs>
              </div>
            </TabsContent>
          )}

          {/* Token Exchange */}
          <TabsContent value="token" className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Successful token exchange returns access and refresh tokens.
            </p>
            <CodeBlock language="json">
{`{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "dGhpcyBpcyBhIHJlZn...",
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "scope": "openid profile email"
}`}
            </CodeBlock>
          </TabsContent>

          {/* Token Refresh */}
          <TabsContent value="refresh" className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Use the refresh token to obtain new access tokens without re-authentication.
            </p>
            <CodeBlock language="bash">
{`curl -X POST /oauth2/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=refresh_token" \\
  -d "refresh_token=YOUR_REFRESH_TOKEN" \\
  -d "client_id=${client.client_id}"${isConfidential ? ' \\\n  -d "client_secret=YOUR_CLIENT_SECRET"' : ''}`}
            </CodeBlock>
          </TabsContent>
        </Tabs>

        <div className="mt-4 pt-4 border-t">
          <p className="text-xs text-muted-foreground">
            <strong>Redirect URI Requirements:</strong> All redirect URIs must use HTTPS in 
            production. HTTP is only allowed for <code>localhost</code> and <code>127.0.0.1</code> during 
            development. Fragment identifiers (#) are not allowed per the OAuth 2.0 specification.
          </p>
        </div>
      </CardContent>
    </Card>
  );
}

function CodeBlock({ children, language }: { children: string; language: string }) {
  return (
    <div className="relative">
      <div className="absolute top-2 right-2 text-xs text-muted-foreground">{language}</div>
      <pre className="bg-muted p-3 rounded-md text-sm overflow-x-auto">
        <code>{children}</code>
      </pre>
    </div>
  );
}
