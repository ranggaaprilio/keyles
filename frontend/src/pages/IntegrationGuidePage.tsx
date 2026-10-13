import { useState } from "react";
import {
  ArrowRight,
  Check,
  CheckCircle2,
  Clipboard,
  Code2,
  ExternalLink,
  KeyRound,
  LockKeyhole,
  RefreshCw,
  Server,
  UserRound,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

const issuer = (
  import.meta.env.VITE_OAUTH_ISSUER ||
  import.meta.env.VITE_API_URL ||
  "http://localhost:8080"
).replace(/\/+$/, "");

const endpoints = [
  ["Authorization", "/oauth2/auth", "Start the browser sign-in flow"],
  ["Token", "/oauth2/token", "Exchange or refresh tokens"],
  ["UserInfo", "/oauth2/userinfo", "Read the signed-in user's claims"],
  ["Revocation", "/oauth2/revoke", "Invalidate an access or refresh token"],
  ["Introspection", "/oauth2/introspect", "Check a token from your backend"],
  [
    "OIDC Discovery",
    "/.well-known/openid-configuration",
    "Discover provider metadata",
  ],
  ["JWKS", "/.well-known/jwks.json", "Verify RS256 token signatures"],
] as const;

const navigation = [
  ["quick-start", "Quick start"],
  ["endpoints", "Endpoints"],
  ["pkce-flow", "PKCE flow"],
  ["tokens", "Use tokens"],
  ["validation", "Validate ID token"],
  ["errors", "Errors"],
  ["production", "Production checklist"],
] as const;

const authorizationCode = `const verifier = generateCodeVerifier();
const challenge = await generateCodeChallenge(verifier);
const state = crypto.randomUUID();

sessionStorage.setItem("pkce_verifier", verifier);
sessionStorage.setItem("oauth_state", state);

const url = new URL("${issuer}/oauth2/auth");
url.searchParams.set("client_id", "YOUR_CLIENT_ID");
url.searchParams.set("redirect_uri", "https://app.example.com/auth/callback");
url.searchParams.set("response_type", "code");
url.searchParams.set("scope", "openid profile email");
url.searchParams.set("state", state);
url.searchParams.set("code_challenge", challenge);
url.searchParams.set("code_challenge_method", "S256");

window.location.href = url.toString();`;

const pkceUtilities = `function generateCodeVerifier(): string {
  const chars =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~";
  const values = new Uint8Array(64);
  crypto.getRandomValues(values);
  return Array.from(values, value => chars[value % chars.length]).join("");
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const digest = await crypto.subtle.digest(
    "SHA-256",
    new TextEncoder().encode(verifier),
  );

  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\\+/g, "-")
    .replace(/\\//g, "_")
    .replace(/=+$/, "");
}`;

const callbackCode = `const callback = new URL(window.location.href);
const error = callback.searchParams.get("error");
const code = callback.searchParams.get("code");
const returnedState = callback.searchParams.get("state");
const storedState = sessionStorage.getItem("oauth_state");

if (error) throw new Error(\`Authorization failed: \${error}\`);
if (!code || returnedState !== storedState) {
  throw new Error("Invalid OAuth callback");
}

const response = await fetch("${issuer}/oauth2/token", {
  method: "POST",
  headers: { "Content-Type": "application/x-www-form-urlencoded" },
  body: new URLSearchParams({
    grant_type: "authorization_code",
    code,
    redirect_uri: "https://app.example.com/auth/callback",
    client_id: "YOUR_CLIENT_ID",
    code_verifier: sessionStorage.getItem("pkce_verifier")!,
  }),
});

if (!response.ok) throw new Error("Token exchange failed");
const tokens = await response.json();

sessionStorage.removeItem("oauth_state");
sessionStorage.removeItem("pkce_verifier");`;

const refreshCode = `const response = await fetch("${issuer}/oauth2/token", {
  method: "POST",
  headers: { "Content-Type": "application/x-www-form-urlencoded" },
  body: new URLSearchParams({
    grant_type: "refresh_token",
    refresh_token: currentRefreshToken,
    client_id: "YOUR_CLIENT_ID",
    // Add client_secret only from a confidential server-side client.
  }),
});

const nextTokens = await response.json();
// Refresh tokens rotate. Replace the old refresh token immediately.`;

const userInfoCode = `curl -H "Authorization: Bearer <access_token>" \\
  ${issuer}/oauth2/userinfo`;

const revokeCode = `curl -X POST ${issuer}/oauth2/revoke \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "token=<refresh_token>" \\
  -d "token_type_hint=refresh_token" \\
  -d "client_id=YOUR_CLIENT_ID"`;

export function IntegrationGuidePage() {
  return (
    <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6">
      <header className="relative overflow-hidden rounded-xl bg-muted p-6 sm:p-8">
        <div className="relative z-10 max-w-3xl">
          <div className="mb-3 flex flex-wrap items-center gap-2">
            <Badge>OAuth 2.0</Badge>
            <Badge variant="outline">OpenID Connect</Badge>
            <Badge variant="outline">PKCE S256</Badge>
          </div>
          <h1 className="text-3xl font-normal leading-[1.1] tracking-[-0.03em] sm:text-5xl">
            Connect your app to Keyles
          </h1>
          <p className="mt-4 max-w-2xl text-base text-muted-foreground">
            A practical guide to registering a client, signing users in, safely
            handling tokens, and preparing your integration for production.
          </p>
          <div className="mt-5 flex flex-wrap gap-2">
            <Button asChild>
              <a href="#quick-start">
                Start integrating <ArrowRight />
              </a>
            </Button>
            <Button asChild variant="outline">
              <a
                href={`${issuer}/.well-known/openid-configuration`}
                target="_blank"
                rel="noreferrer"
              >
                Open discovery <ExternalLink />
              </a>
            </Button>
          </div>
        </div>
      </header>

      <div className="mt-6 grid gap-6 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside className="h-fit overflow-hidden rounded-lg border border-border bg-background lg:sticky lg:top-20">
          <p className="border-b border-border bg-muted px-3 py-2 text-xs font-medium text-muted-foreground">
            On this page
          </p>
          <nav aria-label="Integration guide sections">
            {navigation.map(([id, label], index) => (
              <a
                key={id}
                href={`#${id}`}
                className="flex items-center gap-2 border-b border-border px-3 py-2 text-sm text-foreground no-underline last:border-b-0 hover:bg-muted transition-colors duration-fast"
              >
                <span className="w-5 font-mono text-muted-foreground">
                  {String(index + 1).padStart(2, "0")}
                </span>
                {label}
              </a>
            ))}
          </nav>
        </aside>

        <main className="min-w-0 space-y-8">
          <GuideSection
            id="quick-start"
            number="01"
            title="Quick start"
            description="The shortest path from a registered application to a signed-in user."
          >
            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <QuickStep
                icon={<Code2 />}
                number="1"
                title="Register a client"
                description="Add exact callback URLs in Client Applications and save the client ID."
              />
              <QuickStep
                icon={<LockKeyhole />}
                number="2"
                title="Generate PKCE"
                description="Create a verifier, S256 challenge, state, and optional nonce."
              />
              <QuickStep
                icon={<UserRound />}
                number="3"
                title="Redirect the user"
                description="Send the browser to Keyles for login and consent."
              />
              <QuickStep
                icon={<KeyRound />}
                number="4"
                title="Exchange the code"
                description="Validate state, then exchange the one-time code for tokens."
              />
            </div>

            <div className="mt-4 rounded-md bg-muted p-4">
              <p className="text-xs font-medium">
                Choose the correct client type
              </p>
              <p className="mt-1 text-sm">
                Use a <strong>public client</strong> for SPAs and mobile apps.
                Use a <strong>confidential client</strong> only when a trusted
                backend can keep the client secret out of browser code. PKCE is
                required for both.
              </p>
            </div>
          </GuideSection>

          <GuideSection
            id="endpoints"
            number="02"
            title="Provider endpoints"
            description={`Your configured issuer is ${issuer}. Prefer OIDC discovery instead of hard-coding these URLs.`}
          >
            <div className="overflow-x-auto rounded-lg border border-border">
              <table className="w-full min-w-[680px] border-collapse text-left">
                <thead className="bg-muted text-xs text-muted-foreground">
                  <tr>
                    <th className="px-3 py-2">Endpoint</th>
                    <th className="px-3 py-2">URL</th>
                    <th className="px-3 py-2">Purpose</th>
                  </tr>
                </thead>
                <tbody>
                  {endpoints.map(([name, path, purpose]) => (
                    <tr key={path} className="border-t border-border">
                      <td className="px-3 py-2 text-xs font-medium">
                        {name}
                      </td>
                      <td className="px-3 py-2 font-mono text-xs">
                        {issuer}
                        {path}
                      </td>
                      <td className="px-3 py-2 text-sm">{purpose}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </GuideSection>

          <GuideSection
            id="pkce-flow"
            number="03"
            title="Authorization code flow with PKCE"
            description="Keyles requires the S256 PKCE method. Authorization codes are single-use and expire after five minutes."
          >
            <FlowDiagram />
            <div className="mt-5 grid gap-5 xl:grid-cols-2">
              <CodeBlock
                title="1. PKCE utilities"
                language="TypeScript"
                code={pkceUtilities}
              />
              <CodeBlock
                title="2. Start authorization"
                language="TypeScript"
                code={authorizationCode}
              />
              <div className="xl:col-span-2">
                <CodeBlock
                  title="3. Handle callback and exchange code"
                  language="TypeScript"
                  code={callbackCode}
                />
              </div>
            </div>
            <div className="mt-4 grid gap-3 md:grid-cols-3">
              <ParameterCard
                name="state"
                detail="Required in your app to bind the callback to the browser session and stop CSRF."
              />
              <ParameterCard
                name="nonce"
                detail="Recommended for OIDC. Validate the same value in the returned ID token."
              />
              <ParameterCard
                name="prompt / max_age"
                detail="Optional controls for reauthentication, silent login, consent, and session age."
              />
            </div>
          </GuideSection>

          <GuideSection
            id="tokens"
            number="04"
            title="Use and rotate tokens"
            description="Access tokens last 15 minutes by default. Refresh tokens last seven days and rotate every time they are used."
          >
            <div className="grid gap-4 md:grid-cols-3">
              <TokenCard
                title="Access token"
                lifetime="15 minutes"
                description="RS256 JWT sent as a Bearer token to protected resources."
              />
              <TokenCard
                title="Refresh token"
                lifetime="7 days"
                description="Long-lived credential used once to obtain a fresh token set."
              />
              <TokenCard
                title="ID token"
                lifetime="Token response"
                description="OIDC identity assertion. Validate it before trusting user claims."
              />
            </div>
            <div className="mt-5 grid gap-5 xl:grid-cols-2">
              <CodeBlock
                title="Read user claims"
                language="cURL"
                code={userInfoCode}
              />
              <CodeBlock
                title="Refresh the token set"
                language="TypeScript"
                code={refreshCode}
              />
              <div className="xl:col-span-2">
                <CodeBlock
                  title="Revoke on application logout"
                  language="cURL"
                  code={revokeCode}
                />
              </div>
            </div>
          </GuideSection>

          <GuideSection
            id="validation"
            number="05"
            title="Validate the ID token"
            description="Fetch signing keys through discovery, verify the RS256 signature, then validate every relevant claim."
          >
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {[
                ["iss", `Must exactly match ${issuer}`],
                ["aud", "Must contain your client ID"],
                ["exp", "Must be later than the current time"],
                ["iat", "Should be in the recent past"],
                ["nonce", "Must match when included in your request"],
                ["auth_time", "Check when your app requires fresh login"],
              ].map(([claim, rule]) => (
                <div key={claim} className="rounded-md border border-border p-3">
                  <code className="bg-primary px-2 py-0.5 text-xs text-primary-foreground">
                    {claim}
                  </code>
                  <p className="mt-2 text-sm">{rule}</p>
                </div>
              ))}
            </div>
            <p className="mt-4 rounded-md bg-muted p-4 text-sm text-muted-foreground">
              Use a maintained OpenID Connect or JWT library in production.
              Never treat an access token as proof of authentication, and never
              trust decoded claims before signature and claim validation.
            </p>
          </GuideSection>

          <GuideSection
            id="errors"
            number="06"
            title="Handle protocol errors"
            description="Authorization errors return to the registered callback. Token errors use a JSON response with an error code and safe description."
          >
            <div className="grid gap-4 md:grid-cols-2">
              <ErrorList
                title="Authorization"
                errors={[
                  ["invalid_request", "Missing, malformed, or expired request"],
                  ["invalid_client", "Unknown client ID"],
                  ["access_denied", "User denied consent"],
                  ["login_required", "Silent request needs a login"],
                  ["consent_required", "Silent request needs consent"],
                  ["temporarily_unavailable", "Provider dependency unavailable"],
                ]}
              />
              <ErrorList
                title="Token endpoint"
                errors={[
                  ["invalid_request", "Required token parameter is missing"],
                  ["invalid_grant", "Code expired, reused, or failed PKCE"],
                  ["invalid_client", "Client authentication failed"],
                  ["unsupported_grant_type", "Grant type is not supported"],
                ]}
              />
            </div>
          </GuideSection>

          <GuideSection
            id="production"
            number="07"
            title="Production checklist"
            description="Complete these items before real users authenticate through your integration."
          >
            <div className="grid gap-2 sm:grid-cols-2">
              {[
                "Use HTTPS for the issuer, callback, and every application endpoint.",
                "Register exact redirect URIs. Do not use wildcards or URL fragments.",
                "Keep access tokens in memory for SPAs, never in localStorage.",
                "Keep client secrets only on a trusted server.",
                "Validate state, nonce, signature, issuer, audience, and expiry.",
                "Replace rotated refresh tokens and discard the previous value.",
                "Handle 401 responses and refresh failure by starting login again.",
                "Revoke the refresh token when the user logs out of your app.",
              ].map((item) => (
                <div
                  key={item}
                  className="flex gap-3 rounded-md border border-border bg-background p-3"
                >
                  <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" />
                  <p className="text-sm">{item}</p>
                </div>
              ))}
            </div>
          </GuideSection>
        </main>
      </div>
    </div>
  );
}

function GuideSection({
  id,
  number,
  title,
  description,
  children,
}: {
  id: string;
  number: string;
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <section id={id} className="scroll-mt-20">
      <div className="mb-4 flex gap-3 border-b border-border pb-3">
        <span className="font-mono text-lg font-medium text-muted-foreground">
          {number}
        </span>
        <div>
          <h2 className="text-xl font-medium">
            {title}
          </h2>
          <p className="mt-1 max-w-3xl text-sm text-muted-foreground">
            {description}
          </p>
        </div>
      </div>
      {children}
    </section>
  );
}

function QuickStep({
  icon,
  number,
  title,
  description,
}: {
  icon: React.ReactNode;
  number: string;
  title: string;
  description: string;
}) {
  return (
    <Card className="relative overflow-hidden">
      <span className="absolute right-2 top-1 text-4xl text-border">
        {number}
      </span>
      <CardHeader className="relative pb-2">
        <div className="mb-2 flex h-9 w-9 items-center justify-center rounded-full bg-primary text-primary-foreground">
          {icon}
        </div>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm">{description}</p>
      </CardContent>
    </Card>
  );
}

function FlowDiagram() {
  const actors = [
    [<Code2 key="app" />, "Your app", "PKCE + callback"],
    [<UserRound key="user" />, "User browser", "Login + consent"],
    [<Server key="server" />, "Keyles", "Codes + tokens"],
  ] as const;

  return (
    <div className="grid items-center gap-2 bg-primary p-3 sm:grid-cols-[1fr_auto_1fr_auto_1fr]">
      {actors.map(([icon, title, detail], index) => (
        <div key={title} className="contents">
          <div className="flex items-center gap-3 bg-background p-3">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-border">
              {icon}
            </div>
            <div>
              <p className="text-xs font-medium">{title}</p>
              <p className="text-xs text-muted-foreground">{detail}</p>
            </div>
          </div>
          {index < actors.length - 1 && (
            <ArrowRight className="mx-auto hidden text-primary-foreground sm:block" />
          )}
        </div>
      ))}
    </div>
  );
}

function CodeBlock({
  title,
  language,
  code,
}: {
  title: string;
  language: string;
  code: string;
}) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="min-w-0 rounded-lg border border-border">
      <div className="flex items-center gap-2 border-b border-border bg-muted px-3 py-2">
        <Code2 className="h-4 w-4" />
        <p className="text-xs font-medium">
          {title}
        </p>
        <Badge variant="outline" className="ml-auto hidden sm:inline-flex">
          {language}
        </Badge>
        <button
          type="button"
          onClick={copy}
          className={cn(
            "ml-1 flex items-center gap-1 text-xs font-medium",
            copied ? "text-success" : "text-foreground",
          )}
          aria-label={`Copy ${title}`}
        >
          {copied ? <Check className="h-3.5 w-3.5" /> : <Clipboard className="h-3.5 w-3.5" />}
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
      <pre className="max-h-[420px] overflow-auto rounded-lg bg-primary p-4 font-mono text-xs leading-5 text-primary-foreground">
        <code>{code}</code>
      </pre>
    </div>
  );
}

function ParameterCard({ name, detail }: { name: string; detail: string }) {
  return (
    <div className="rounded-md border border-border p-3">
      <p className="font-mono text-xs font-medium">{name}</p>
      <p className="mt-1 text-sm text-muted-foreground">{detail}</p>
    </div>
  );
}

function TokenCard({
  title,
  lifetime,
  description,
}: {
  title: string;
  lifetime: string;
  description: string;
}) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <CardTitle>{title}</CardTitle>
          <RefreshCw className="h-4 w-4 shrink-0" />
        </div>
        <CardDescription>{lifetime}</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm">{description}</p>
      </CardContent>
    </Card>
  );
}

function ErrorList({
  title,
  errors,
}: {
  title: string;
  errors: ReadonlyArray<readonly [string, string]>;
}) {
  return (
    <div className="border-t border-border">
      <h3 className="bg-primary px-3 py-2 text-xs font-medium text-primary-foreground">
        {title}
      </h3>
      <dl>
        {errors.map(([code, detail]) => (
          <div
            key={code}
            className="grid gap-1 border-t border-border px-3 py-2 first:border-t-0 sm:grid-cols-[170px_1fr]"
          >
            <dt>
              <code className="text-xs font-medium">{code}</code>
            </dt>
            <dd className="text-sm text-muted-foreground">{detail}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
