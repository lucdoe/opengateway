# Auth

Add `auth` to an endpoint's `Plugins`, or a service's for all its endpoints, to require a `Bearer` JWT. A token is accepted when:

- it is signed with `SecretKey` or with a key from [`JWKSURL`](#tokens-from-an-identity-provider), and carries `exp` (`401` otherwise)
- its `iss` and `aud` match `Issuer` and `Audience`, when they are set (`401` otherwise)
- its space-separated `scope` claim has every scope in `Plugins.JWT.Scope`, and in the endpoint's `Scope` (`403` otherwise)

```yaml
Plugins:
  JWT:
    SecretKey: "${JWT_SECRET}"
    Issuer: "https://auth.example.com/"
    Audience: "my-api"
    Scope: "" # scopes every token needs
Services:
  accounts:
    URL: "https://accounts.internal"
    Plugins: ["auth"]
    Endpoints:
      - HTTPMethod: "DELETE"
        Path: "/users/{id}"
        Scope: "users:admin" # this endpoint also needs users:admin
```

With `SecretKey`, `SigningMethod` can be `HS256` (the default), `HS384` or `HS512`. A `401` never says why the token was rejected.

## Tokens from an identity provider

To accept tokens from Auth0, Keycloak, Cognito, Entra ID or any other OpenID Connect provider, set `JWKSURL` to the provider's key set (its `jwks_uri`) instead of `SecretKey`:

```yaml
Plugins:
  JWT:
    JWKSURL: "https://auth.example.com/.well-known/jwks.json"
    Issuer: "https://auth.example.com/"
    Audience: "my-api"
```

- **Algorithms:** RS256/384/512, PS256/384/512, ES256/384/512 and EdDSA are accepted, each only with a key of its type. `SigningMethod` limits them to one. HS algorithms never are, so a public key can't be used as an HMAC secret.
- **Key refresh:** the keys are fetched at startup, then every hour. A token naming an unknown `kid` fetches them again right away, at most once a minute, so rotated keys work without a restart.
- **When the provider is down:** if it can't be reached at startup, the gateway still starts, retries every 30 seconds, and rejects tokens until it has keys. A failed refresh keeps the known keys.
- **https:** `JWKSURL` must use https, except to hosts inside your network, such as a Compose service.

## Warnings

Startup and `check` warn about auth settings that are valid but risky. They don't stop the gateway or fail `check`.

| Warning | Why |
| --- | --- |
| `SecretKey` is the example secret, or shorter than 32 bytes | Anyone can forge tokens with a known secret. HS256 needs 256 bits to resist guessing. |
| `Issuer` or `Audience` is empty | Tokens from other issuers, or meant for other services or apps, are accepted. |
| A service with auth endpoints has a plain `http://` URL to an outside host | The token is forwarded unencrypted. Loopback and private addresses, names without a dot (like a Compose service) and `.local`, `.internal`, `.lan`, `.svc` and `.localhost` names count as inside. |
